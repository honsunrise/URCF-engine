package processes

import (
	"errors"
	"os"
	"syscall"
	"strings"
	"sync"
	"github.com/zhsyourai/URCF-engine/services/processes/watchdog"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	Start(name string, workDir string, cmd string,
		args []string, env map[string]string) (*Process, error)
	FindByName(name string) *Process
	Stop(s *Process) error
	Restart(s *Process) (*Process, error)
	Kill(s *Process) error
	Clean(s *Process) error
	Watch(s *Process) error
	IsAlive(s *Process) bool
}

// pluginService is a os.pluginService wrapper with Statistics and more info that will be used on Master to maintain
// the process health.
type pluginService struct {
	procMap  sync.Map
	watchDog watchdog.WatchDog
}

func NewPluginService() Service {
	pluginService := &pluginService{
		watchDog: watchdog.NewWatcherDog(),
	}
	go pluginService.runAutoStart()
	return pluginService
}

func buildEnv(env map[string]string) (result []string) {
	for k, v := range env {
		result = append(result, k+"="+v)
	}
	return
}

func (s *pluginService) runAutoStart() {
	for proc := range s.watchDog.GetDeathsChan() {
		if !proc.KeepAlive {
			master.Lock()
			master.updateStatus(proc)
			master.Unlock()
			log.Infof("Proc %s does not have keep alive set. Will not be restarted.", proc.Identifier())
			continue
		}
		log.Infof("Restarting proc %s.", proc.Identifier())
		if proc.IsAlive() {
			log.Warnf("Proc %s was supposed to be dead, but it is alive.", proc.Identifier())
		}
		master.Lock()
		err := master.restart(proc)
		master.Unlock()
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Identifier(), err)
		}
	}
}

func (s *pluginService) Start(name string, workDir string, cmd string,
	args []string, env map[string]string) (proc *Process, err error) {
	proc = &Process{
		Name:    name,
		Cmd:     cmd,
		Args:    args,
		Env:     env,
		WorkDir: workDir,
	}
	rStdIn, lStdOut, err := os.Pipe()
	if err != nil {
		return
	}
	proc.StdIn = lStdOut
	lStdIn, rStdOut, err := os.Pipe()
	if err != nil {
		return
	}
	proc.StdOut = lStdIn
	lErrIn, rErrOut, err := os.Pipe()
	if err != nil {
		return
	}
	proc.StdErr = lErrIn
	finalEnv := make(map[string]string)
	for _, e := range os.Environ() {
		es := strings.Split(e, "=")
		finalEnv[es[0]] = es[1]
	}
	for k, v := range env {
		finalEnv[k] = v
	}
	procAtr := &os.ProcAttr{
		Dir: workDir,
		Env: buildEnv(finalEnv),
		Files: []*os.File{
			rStdIn,
			rStdOut,
			rErrOut,
		},
	}
	finalArgs := append([]string{cmd}, args...)
	process, err := os.StartProcess(cmd, finalArgs, procAtr)
	if err != nil {
		return
	}
	proc.Process = process
	proc.Pid = proc.Process.Pid
	proc.Statistics.InitUpTime()
	s.procMap.Store(name, proc)
	return
}

func (s *pluginService) FindByName(name string) *Process {
	if p, ok := s.procMap.Load(name); ok {
		return p.(*Process)
	}
	return nil
}

func (s *pluginService) Stop(p *Process) error {
	if s.FindByName(p.Name) == nil || p.Process == nil {
		return errors.New("process does not exist")
	}
	defer p.Process.Release()
	err := p.Process.Signal(syscall.SIGTERM)
	return err
}

func (s *pluginService) Restart(p *Process) (proc *Process, err error) {
	if s.IsAlive(p) {
		err := s.Stop(p)
		if err != nil {
			return nil, err
		}
	}
	return s.Start(p.Name, p.WorkDir, p.Cmd, p.Args, p.Env)
}

func (s *pluginService) Kill(p *Process) error {
	if s.FindByName(p.Name) == nil || p.Process == nil {
		return errors.New("process does not exist")
	}
	defer p.Process.Release()
	err := p.Process.Signal(syscall.SIGKILL)
	return err
}

func (s *pluginService) Clean(p *Process) error {
	p.Process.Release()
	return os.RemoveAll(p.WorkDir)
}

func (s *pluginService) Watch(p *Process) error {
	return s.watchDog.StartWatch(p)
}

func (s *pluginService) IsAlive(p *Process) bool {
	if s.FindByName(p.Name) == nil || p.Process == nil {
		return false
	}
	_, err := os.FindProcess(p.Pid)
	if err != nil {
		return false
	}
	return p.Process.Signal(syscall.Signal(0)) == nil
}
