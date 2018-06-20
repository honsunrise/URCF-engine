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

var ProcessExist = errors.New("process exist")
var ProcessNotExist = errors.New("process not exist")
var ProcessNotRun = errors.New("process does not run")

type Service interface {
	services.ServiceLifeCycle
	Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
		option models.ProcessOption) (*types.Process, error)
	ListAll() []*types.Process
	FindByName(name string) *types.Process
	Start(name string) error
	Stop(name string) error
	Restart(name string) error
	Kill(name string) error
	Clean(name string) error
	Watch(name string) error
	Wait(name string) <-chan error
	IsAlive(name string) bool
}

type processPair struct {
	proc         *types.Process
	procAttr     *os.ProcAttr
	finalArgs    []string
	lock         sync.Mutex
	ExitingChan  chan struct{}
	ExitDoneChan chan struct{}
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
		if proc.Option&models.AutoRestart != 0 {
			log.Infof("process %s does not have keep alive set. Will not be restarted.", proc.Name)
			continue
		}
		log.Infof("Restarting process %s.", proc.Name)
		if s.IsAlive(proc.Name) {
			log.Warnf("process %s was supposed to be dead, but it is alive.", proc.Name)
		}

		err := s.Restart(proc.Name)
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Name, err)
		}
	}
}

func (s *processesService) ListAll() (processes []*types.Process) {
	processes = []*types.Process{}
	s.procMap.Range(func(key, value interface{}) bool {
		processPair := value.(*processPair)
		processes = append(processes, processPair.proc)
		return true
	})
	return
}

func (s *processesService) Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
	option models.ProcessOption) (*types.Process, error) {
	var err error
	_, loaded := s.procMap.Load(name)
	if loaded {
		return nil, ProcessExist
	}
	proc := &types.Process{
		ProcessParam: models.ProcessParam{
			Name:    name,
			Cmd:     cmd,
			Args:    args,
			Env:     env,
			WorkDir: workDir,
			Option:  option,
		},
	}
	rStdIn, lStdIn, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	proc.StdIn = lStdIn
	lStdOut, rStdOut, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	proc.StdOut = lStdOut
	lStdErr, rStdErr, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	proc.StdErr = lStdErr
	lDataOut, rDataOut, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	proc.DataOut = lDataOut
	if option&models.HookLog != 0 {
		err = logservice.GetInstance().WarpReader(name, lStdErr)
		if err != nil {
			return nil, err
		}
		err = logservice.GetInstance().WarpReader(name, lStdOut)
		if err != nil {
			return nil, err
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

	pp := &processPair{
		proc:         proc,
		procAttr:     procAttr,
		finalArgs:    append([]string{cmd}, args...),
		ExitingChan:  make(chan struct{}),
		ExitDoneChan: make(chan struct{}),
	}

	go func() {
		<-pp.ExitingChan
		proc.StdIn.Close()
		proc.StdOut.Close()
		proc.DataOut.Close()
		proc.StdErr.Close()
		for _, f := range procAttr.Files {
			f.Close()
		}
		s.procMap.Delete(name)
		proc.Status = types.Exited
		close(pp.ExitDoneChan)
	}()

	proc.Status = types.Prepare

	_, loaded = s.procMap.LoadOrStore(name, pp)
	if loaded {
		close(pp.ExitingChan)
		return nil, ProcessExist
	}

	return proc, err
}

func (s *processesService) FindByName(name string) *types.Process {
	if p, ok := s.procMap.Load(name); ok {
		return p.(*processPair).proc
	}
	return nil
}

func (s *processesService) Start(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	process, err := os.StartProcess(pp.proc.Cmd, pp.finalArgs, pp.procAttr)
	if err != nil {
		return err
	}

	go func() {
		pp.proc.Process.Wait()
		close(pp.ExitingChan)
	}()

	if pp.proc.Option&models.AutoRestart != 0 {
		err = s.watchDog.StartWatch(pp.proc)
		if err != nil {
			return err
		}
	}

	pp.proc.Process = process
	pp.proc.Pid = process.Pid
	pp.proc.Statistics.InitUpTime()
	pp.proc.Status = types.Running
	return nil
}

func (s *processesService) Stop(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	if pp.proc.Status != types.Running {
		return ProcessNotRun
	}

	pp.proc.Status = types.Exiting

	defer pp.proc.Process.Release()
	return pp.proc.Process.Signal(syscall.SIGTERM)
}

func (s *processesService) Restart(name string) error {
	if s.IsAlive(name) {
		err := s.Stop(name)
		if err != nil {
			return err
		}
	}
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	<-pp.ExitDoneChan
	_, err := s.Prepare(pp.proc.Name, pp.proc.WorkDir, pp.proc.Cmd, pp.proc.Args, pp.proc.Env, pp.proc.Option)
	if err != nil {
		return err
	}
	return s.Start(name)
}

func (s *processesService) Kill(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	if pp.proc.Status != types.Running {
		return ProcessNotRun
	}

	pp.proc.Status = types.Exiting

	defer pp.proc.Process.Release()
	return pp.proc.Process.Kill()
}

func (s *processesService) Clean(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	if pp.proc.Status == types.Running {
		s.Stop(name)
	}

	pp.proc.Process.Release()
	return os.RemoveAll(pp.proc.WorkDir)
}

func (s *processesService) Watch(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	return s.watchDog.StartWatch(pp.proc)
}

func (s *processesService) Wait(name string) <-chan error {
	ret := make(chan error, 1)
	result, ok := s.procMap.Load(name)
	if !ok {
		ret <- ProcessNotExist
		return ret
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	go func() {
		<-pp.ExitDoneChan
		close(ret)
	}()

	return ret
}

func (s *processesService) IsAlive(name string) bool {
	result, ok := s.procMap.Load(name)
	if !ok {
		return false
	}
	pp := result.(*processPair)
	pp.lock.Lock()
	defer pp.lock.Unlock()

	_, err := os.FindProcess(pp.proc.Pid)
	if err != nil {
		return false
	}
	return pp.proc.Process.Signal(syscall.Signal(0)) == nil
}
