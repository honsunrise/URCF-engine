package processes

import (
	"errors"
	"os"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/processes/types"
	"github.com/zhsyourai/URCF-engine/services/processes/watchdog"
)

type Service interface {
	services.ServiceLifeCycle
	Start(name string, workDir string, cmd string,
		args []string, env map[string]string, option models.ProcessOption) (*types.Process, error)
	FindByName(name string) *types.Process
	Stop(s *types.Process) error
	Restart(s *types.Process) (*types.Process, error)
	Kill(s *types.Process) error
	Clean(s *types.Process) error
	Watch(s *types.Process) error
	IsAlive(s *types.Process) bool
}

// processesService is a os.processesService wrapper with Statistics and more info that will be used on Master to maintain
// the process health.
type processesService struct {
	services.InitHelper
	procMap  sync.Map
	watchDog watchdog.Service
}

func (s *processesService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *processesService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

var instance *processesService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &processesService{
			watchDog: watchdog.GetInstance(),
		}
		go instance.runAutoReStart()
	})
	return instance
}

func buildEnv(env map[string]string) (result []string) {
	for k, v := range env {
		result = append(result, k+"="+v)
	}
	return
}

func (s *processesService) runAutoReStart() {
	for proc := range s.watchDog.GetDeathsChan() {
		if !proc.KeepAlive {
			log.Infof("process %s does not have keep alive set. Will not be restarted.", proc.Name)
			continue
		}
		log.Infof("Restarting process %s.", proc.Name)
		if s.IsAlive(proc) {
			log.Warnf("process %s was supposed to be dead, but it is alive.", proc.Name)
		}

		_, err := s.Restart(proc)
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Name, err)
		}
	}
}

func (s *processesService) Start(name string, workDir string, cmd string,
	args []string, env map[string]string, option models.ProcessOption) (proc *types.Process, err error) {
	proc = &types.Process{
		ProcessParam: models.ProcessParam{
			Name:    name,
			Cmd:     cmd,
			Args:    args,
			Env:     env,
			WorkDir: workDir,
			Option:  option,
		},
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

	if option&models.AutoRestart != 0 {
		err = s.watchDog.StartWatch(proc)
		if err != nil {
			return
		}
	}

	proc.Process = process
	proc.Pid = proc.Process.Pid
	proc.Statistics.InitUpTime()
	s.procMap.Store(name, proc)
	return
}

func (s *processesService) FindByName(name string) *types.Process {
	if p, ok := s.procMap.Load(name); ok {
		return p.(*types.Process)
	}
	return nil
}

func (s *processesService) Stop(p *types.Process) error {
	if s.FindByName(p.Name) == nil || p.Process == nil {
		return errors.New("process does not exist")
	}
	defer p.Process.Release()
	err := p.Process.Signal(syscall.SIGTERM)
	return err
}

func (s *processesService) Restart(p *types.Process) (proc *types.Process, err error) {
	if s.IsAlive(p) {
		err := s.Stop(p)
		if err != nil {
			return nil, err
		}
	}
	return s.Start(p.Name, p.WorkDir, p.Cmd, p.Args, p.Env, p.Option)
}

func (s *processesService) Kill(p *types.Process) error {
	if s.FindByName(p.Name) == nil || p.Process == nil {
		return errors.New("process does not exist")
	}
	defer p.Process.Release()
	err := p.Process.Signal(syscall.SIGKILL)
	return err
}

func (s *processesService) Clean(p *types.Process) error {
	p.Process.Release()
	return os.RemoveAll(p.WorkDir)
}

func (s *processesService) Watch(p *types.Process) error {
	return s.watchDog.StartWatch(p)
}

func (s *processesService) IsAlive(p *types.Process) bool {
	if s.FindByName(p.Name) == nil || p.Process == nil {
		return false
	}
	_, err := os.FindProcess(p.Pid)
	if err != nil {
		return false
	}
	return p.Process.Signal(syscall.Signal(0)) == nil
}
