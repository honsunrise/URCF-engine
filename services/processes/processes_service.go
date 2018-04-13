package processes

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services"
	logservice "github.com/zhsyourai/URCF-engine/services/log"
	"github.com/zhsyourai/URCF-engine/services/processes/types"
	"github.com/zhsyourai/URCF-engine/services/processes/watchdog"
	"os"
	"strings"
	"sync"
	"syscall"
)

type Service interface {
	services.ServiceLifeCycle
	Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
		option models.ProcessOption) (*types.Process, error)
	FindByName(name string) *types.Process
	Start(p *types.Process) error
	Stop(p *types.Process) error
	Restart(p *types.Process) error
	Kill(p *types.Process) error
	Clean(p *types.Process) error
	Watch(p *types.Process) error
	IsAlive(p *types.Process) bool
}

type processPair struct {
	proc      *types.Process
	procAttr  *os.ProcAttr
	finalArgs []string
	lock      sync.Mutex
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

		err := s.Restart(proc)
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Name, err)
		}
	}
}

func (s *processesService) loadProcessPair(proc *types.Process) (*processPair, error) {
	pp, ok := s.procMap.Load(proc.Name)
	if !ok || pp.(*processPair).proc != proc {
		return nil, errors.New("process does not exist")
	}
	return pp.(*processPair), nil
}

func (s *processesService) Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
	option models.ProcessOption) (proc *types.Process, err error) {
	proc = &types.Process{
		ProcessParam: models.ProcessParam{
			Name:    name,
			Cmd:     cmd,
			Args:    args,
			Env:     env,
			WorkDir: workDir,
			Option:  option,
		},
		ExitChan: make(chan struct{}),
	}
	rStdIn, lStdIn, err := os.Pipe()
	if err != nil {
		return
	}
	proc.StdIn = lStdIn
	lStdOut, rStdOut, err := os.Pipe()
	if err != nil {
		return
	}
	proc.StdOut = lStdOut
	lStdErr, rStdErr, err := os.Pipe()
	if err != nil {
		return
	}
	proc.StdErr = lStdErr
	lDataOut, rDataOut, err := os.Pipe()
	if err != nil {
		return
	}
	proc.DataOut = lDataOut
	if option&models.HookLog != 0 {
		err = logservice.GetInstance().WarpReader(name, lStdErr)
		if err != nil {
			return
		}
		err = logservice.GetInstance().WarpReader(name, lStdOut)
		if err != nil {
			return
		}
	}

	finalEnv := make(map[string]string)
	for _, e := range os.Environ() {
		es := strings.Split(e, "=")
		finalEnv[es[0]] = es[1]
	}
	for k, v := range env {
		finalEnv[k] = v
	}
	procAttr := &os.ProcAttr{
		Dir: workDir,
		Env: buildEnv(finalEnv),
		Files: []*os.File{
			rStdIn,
			rStdOut,
			rStdErr,
			rDataOut,
		},
	}

	go func() {
		<-proc.ExitChan
		proc.StdIn.Close()
		proc.StdOut.Close()
		proc.DataOut.Close()
		proc.StdErr.Close()
		for _, f := range procAttr.Files {
			f.Close()
		}
		proc.Status = types.Exited
	}()

	proc.Status = types.Prepare

	_, loaded := s.procMap.LoadOrStore(name, &processPair{
		proc:      proc,
		procAttr:  procAttr,
		finalArgs: append([]string{cmd}, args...),
	})
	if loaded {
		return nil, errors.New("process exist")
	}

	return
}

func (s *processesService) FindByName(name string) *types.Process {
	if p, ok := s.procMap.Load(name); ok {
		return p.(*processPair).proc
	}
	return nil
}

func (s *processesService) Start(proc *types.Process) error {
	pp, err := s.loadProcessPair(proc)
	if err != nil {
		return err
	}
	pp.lock.Lock()
	defer pp.lock.Unlock()

	process, err := os.StartProcess(proc.Cmd, pp.finalArgs, pp.procAttr)
	if err != nil {
		return err
	}

	go func() {
		proc.Process.Wait()
		close(proc.ExitChan)
	}()

	if proc.Option&models.AutoRestart != 0 {
		err = s.watchDog.StartWatch(proc)
		if err != nil {
			return err
		}
	}

	proc.Process = process
	proc.Pid = proc.Process.Pid
	proc.Statistics.InitUpTime()
	proc.Status = types.Running
	return nil
}

func (s *processesService) Stop(proc *types.Process) error {
	pp, err := s.loadProcessPair(proc)
	if err != nil {
		return err
	}
	pp.lock.Lock()
	defer pp.lock.Unlock()

	if proc.Status != types.Running {
		return errors.New("process does not run")
	}

	proc.Status = types.Exiting

	defer proc.Process.Release()
	err = proc.Process.Signal(syscall.SIGTERM)
	return err
}

func (s *processesService) Restart(proc *types.Process) error {
	if s.IsAlive(proc) {
		err := s.Stop(proc)
		if err != nil {
			return err
		}
	}
	<-proc.ExitChan
	process, err := s.Prepare(proc.Name, proc.WorkDir, proc.Cmd, proc.Args, proc.Env, proc.Option)
	if err != nil {
		return err
	}
	return s.Start(process)
}

func (s *processesService) Kill(proc *types.Process) error {
	pp, err := s.loadProcessPair(proc)
	if err != nil {
		return err
	}
	pp.lock.Lock()
	defer pp.lock.Unlock()

	if proc.Status != types.Running {
		return errors.New("process does not run")
	}

	proc.Status = types.Exiting

	defer proc.Process.Release()
	err = proc.Process.Kill()
	return err
}

func (s *processesService) Clean(proc *types.Process) error {
	pp, err := s.loadProcessPair(proc)
	if err != nil {
		return err
	}
	pp.lock.Lock()
	defer pp.lock.Unlock()

	if proc.Status == types.Running {
		s.Stop(proc)
	}

	proc.Process.Release()
	return os.RemoveAll(proc.WorkDir)
}

func (s *processesService) Watch(proc *types.Process) error {
	return s.watchDog.StartWatch(proc)
}

func (s *processesService) IsAlive(proc *types.Process) bool {
	_, err := s.loadProcessPair(proc)
	if err != nil {
		return false
	}
	_, err = os.FindProcess(proc.Pid)
	if err != nil {
		return false
	}
	return proc.Process.Signal(syscall.Signal(0)) == nil
}
