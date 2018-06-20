package processes

import (
	"errors"
	"github.com/looplab/fsm"
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
var ProcessStillRun = errors.New("process still run")

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
	Remove(name string) error
	Clean(name string) error
	Watch(name string) error
	Wait(name string) <-chan error
	IsAlive(name string) bool
}

type processPair struct {
	proc         *types.Process
	procAttr     *os.ProcAttr
	finalArgs    []string
	ExitDoneChan chan struct{}
	FSM          *fsm.FSM
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
		if proc.Option&models.AutoRestart == 0 {
			log.Infof("process %s does not have keep alive set. Will not be restarted.", proc.Name)
			continue
		}
		log.Infof("Restarting process %s.", proc.Name)
		if proc.State == types.Running {
			log.Warnf("process %s was supposed to be dead, but it is alive.", proc.Name)
		}

		err := s.Restart(proc.Name)
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Name, err)
		}
	}
}

func (s *processesService) init(pp *processPair) error {
	rStdIn, lStdIn, err := os.Pipe()
	proc := pp.proc
	if err != nil {
		return err
	}
	proc.StdIn = lStdIn
	lStdOut, rStdOut, err := os.Pipe()
	if err != nil {
		return err
	}
	proc.StdOut = lStdOut
	lStdErr, rStdErr, err := os.Pipe()
	if err != nil {
		return err
	}
	proc.StdErr = lStdErr
	lDataOut, rDataOut, err := os.Pipe()
	if err != nil {
		return err
	}
	proc.DataOut = lDataOut
	if proc.Option&models.HookLog != 0 {
		err = logservice.GetInstance().WarpReader(proc.Name, lStdErr)
		if err != nil {
			return err
		}
		err = logservice.GetInstance().WarpReader(proc.Name, lStdOut)
		if err != nil {
			return err
		}
	}

	finalEnv := make(map[string]string)
	for _, e := range os.Environ() {
		es := strings.Split(e, "=")
		finalEnv[es[0]] = es[1]
	}
	for k, v := range proc.Env {
		finalEnv[k] = v
	}
	procAttr := &os.ProcAttr{
		Dir: proc.WorkDir,
		Env: buildEnv(finalEnv),
		Files: []*os.File{
			rStdIn,
			rStdOut,
			rStdErr,
			rDataOut,
		},
	}
	pp.procAttr = procAttr
	pp.finalArgs = append([]string{proc.Cmd}, proc.Args...)
	pp.ExitDoneChan = make(chan struct{}, 1)
	return proc.State.Enter(types.Prepare)
}

func (s *processesService) release(pp *processPair) {
	pp.proc.StdIn.Close()
	pp.proc.StdOut.Close()
	pp.proc.DataOut.Close()
	pp.proc.StdErr.Close()
	for _, f := range pp.procAttr.Files {
		f.Close()
	}
	pp.proc.Statistics.AddStop()
	pp.proc.Statistics.SetLastStopTime()
	if pp.proc.Process != nil {
		pp.proc.Process.Release()
	}
	close(pp.ExitDoneChan)
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
	_, loaded := s.procMap.Load(name)
	if loaded {
		return nil, ProcessExist
	}

	pp := &processPair{
		proc: &types.Process{
			ProcessParam: models.ProcessParam{
				Name:    name,
				Cmd:     cmd,
				Args:    args,
				Env:     env,
				WorkDir: workDir,
				Option:  option,
			},
			State: types.Exited,
		},
		FSM: fsm.NewFSM(
			types.Exited.String(),
			fsm.Events{
				{Name: "start", Src: []string{"closed"}, Dst: "open"},
				{Name: "stop", Src: []string{"open"}, Dst: "closed"},
				{Name: "restart", Src: []string{"open"}, Dst: "closed"},
				{Name: "kill", Src: []string{"open"}, Dst: "closed"},
			},
			fsm.Callbacks{
				"enter_state": func(e *fsm.Event) { d.enterState(e) },
			},
		),
	}

	_, loaded = s.procMap.LoadOrStore(name, pp)
	if loaded {
		return nil, ProcessExist
	}

	err := s.init(pp)
	if err != nil {
		return nil, err
	}

	return pp.proc, nil
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

	if err := pp.proc.State.CanEnter(types.Running); err == nil {
		process, err := os.StartProcess(pp.proc.Cmd, pp.finalArgs, pp.procAttr)
		if err != nil {
			return err
		}

		pp.proc.Process = process
		pp.proc.Pid = process.Pid
		pp.proc.Statistics.InitStartUpTime()
		pp.proc.State = types.Running

		go func() {
			pp.proc.Process.Wait()
			if pp.proc.State.Is(types.Exiting) {
				s.release(pp)
				if err := pp.proc.State.Enter(types.Exited); err != nil {
					panic(err)
				}
			}
		}()

		if pp.proc.Option&models.AutoRestart != 0 {
			err = s.watchDog.StartWatch(pp.proc)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *processesService) Stop(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	if err := pp.proc.State.CanEnter(types.Exiting); err == nil {
		err = pp.proc.Process.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}
		return pp.proc.State.Enter(types.Exiting)
	} else {
		return err
	}
}

func (s *processesService) Restart(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	err := s.Stop(name)
	if err != nil {
		return err
	}
	<-pp.ExitDoneChan
	err = s.init(pp)
	if err != nil {
		return err
	}
	err = s.Start(name)
	if err != nil {
		return err
	}
	pp.proc.Statistics.AddRestart()
	return nil
}

func (s *processesService) Kill(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	if err := pp.proc.State.CanEnter(types.Exiting); err == nil {
		err = pp.proc.Process.Kill()
		if err != nil {
			return err
		}
		return pp.proc.State.Enter(types.Exiting)
	} else {
		return err
	}
}

func (s *processesService) Clean(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	err := s.Stop(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(pp.proc.WorkDir)
}

func (s *processesService) Remove(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	if pp.proc.State == types.Running {
		return ProcessStillRun
	}

	s.procMap.Delete(name)
	return nil
}

func (s *processesService) Watch(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

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

	_, err := os.FindProcess(pp.proc.Pid)
	if err != nil {
		return false
	}
	return pp.proc.Process.Signal(syscall.Signal(0)) == nil
}
