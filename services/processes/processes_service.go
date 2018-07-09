package processes

import (
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services"
	logservice "github.com/zhsyourai/URCF-engine/services/log"
	"github.com/zhsyourai/URCF-engine/services/processes/watchdog"
)

var ProcessExist = errors.New("process exist")
var ProcessNotExist = errors.New("process not exist")
var ProcessNotRun = errors.New("process does not run")
var ProcessStillRun = errors.New("process still run")
var OperatorNotComplete = errors.New("current have operator not complete")

type Service interface {
	services.ServiceLifeCycle
	Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
		option models.ProcessOption) (*models.Process, error)
	ListAll() []*models.Process
	FindByName(name string) *models.Process
	Start(name string) error
	Stop(name string) error
	Restart(name string) error
	Kill(name string) error
	Remove(name string) error
	Watch(name string) error
	Wait(name string) <-chan error
	WaitRestart(name string) <-chan error
	IsAlive(name string) bool
}

type processPair struct {
	proc             *models.Process
	procAttr         *os.ProcAttr
	finalArgs        []string
	osProcState      *os.ProcessState
	FSM              *fsm.FSM
	exitWaitGroup    sync.WaitGroup
	restartWaitGroup sync.WaitGroup
	firstRun         atomic.Value
}

func (pp *processPair) Wait() *os.ProcessState {
	pp.exitWaitGroup.Wait()
	return pp.osProcState
}

func (pp *processPair) done(state *os.ProcessState) {
	pp.osProcState = state
	pp.exitWaitGroup.Done()
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
		result, ok := s.procMap.Load(proc.Name)
		if !ok {
			log.Infof("process %s not exist. Will not be restarted.", proc.Name)
			continue
		}
		pp := result.(*processPair)
		if proc.Option&models.AutoRestart == 0 {
			log.Infof("process %s does not have AutoRestart set. Will not be restarted.", proc.Name)
			continue
		}
		log.Infof("Restarting process %s.", proc.Name)
		if proc.State == models.Running {
			log.Warnf("process %s was supposed to be dead, but it is alive.", proc.Name)
		}

		err := pp.FSM.Event("init")
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Name, err)
			continue
		}

		err = pp.FSM.Event("start")
		if err != nil {
			log.Warnf("Could not restart process %s due to %s.", proc.Name, err)
			continue
		}
	}
}

func (s *processesService) init(name string) error {
	result, loaded := s.procMap.Load(name)
	if !loaded {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	needRelease := true
	defer func() {
		if needRelease {
			s.release(pp)
		}
	}()
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
	pp.exitWaitGroup.Add(1)
	pp.restartWaitGroup.Add(1)
	needRelease = false
	return nil
}

func (s *processesService) start(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	process, err := os.StartProcess(pp.proc.Cmd, pp.finalArgs, pp.procAttr)
	if err != nil {
		return err
	}

	pp.proc.Process = process
	pp.proc.Pid = process.Pid
	pp.proc.Statistics.InitStartUpTime()

	if pp.proc.Option&models.AutoRestart != 0 {
		err = s.watchDog.StartWatch(pp.proc, pp)
		if err != nil {
			return err
		}
	}

	go func() {
		state, _ := pp.proc.Process.Wait()
		pp.FSM.Event("stopDone", state)
	}()

	return nil
}

func (s *processesService) stop(name string, isKill bool) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)
	_, err := os.FindProcess(pp.proc.Pid)
	if err == nil && pp.proc.Process.Signal(syscall.Signal(0)) == nil {
		if isKill {
			err := pp.proc.Process.Kill()
			if err != nil {
				return err
			}
		} else {
			err := pp.proc.Process.Signal(syscall.SIGTERM)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *processesService) release(pp *processPair) {
	if pp.proc.StdIn != nil {
		pp.proc.StdIn.Close()
	}
	if pp.proc.StdOut != nil {
		pp.proc.StdOut.Close()
	}
	if pp.proc.StdErr != nil {
		pp.proc.StdErr.Close()
	}
	if pp.proc.DataOut != nil {
		pp.proc.DataOut.Close()
	}
	for _, f := range pp.procAttr.Files {
		f.Close()
	}
	if pp.proc.Process != nil {
		pp.proc.Process.Release()
	}
}

func (s *processesService) ListAll() (processes []*models.Process) {
	processes = []*models.Process{}
	s.procMap.Range(func(key, value interface{}) bool {
		processPair := value.(*processPair)
		processes = append(processes, processPair.proc)
		return true
	})
	return
}

func (s *processesService) Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
	option models.ProcessOption) (*models.Process, error) {
	_, loaded := s.procMap.Load(name)
	if loaded {
		return nil, ProcessExist
	}

	var pp *processPair
	pp = &processPair{
		proc: &models.Process{
			ProcessParam: models.ProcessParam{
				Name:    name,
				Cmd:     cmd,
				Args:    args,
				Env:     env,
				WorkDir: workDir,
				Option:  option,
			},
			State: models.Exited,
		},
		FSM: fsm.NewFSM(
			"exited",
			fsm.Events{
				{Name: "init", Src: []string{"exited"}, Dst: "prepare"},
				{Name: "start", Src: []string{"prepare"}, Dst: "running"},
				{Name: "stop", Src: []string{"running"}, Dst: "exiting"},
				{Name: "stopDone", Src: []string{"running", "exiting"}, Dst: "exited"},
				{Name: "remove", Src: []string{"exited"}, Dst: "removed"},
			},
			fsm.Callbacks{
				"before_init": func(e *fsm.Event) {
					err := s.init(name)
					if err != nil {
						e.Cancel(err)
					}
				},
				"enter_prepare": func(e *fsm.Event) {
					pp.proc.State = models.Prepare
				},
				"before_start": func(e *fsm.Event) {
					err := s.start(name)
					if err != nil {
						e.Cancel(err)
					}
				},
				"enter_running": func(e *fsm.Event) {
					if !pp.firstRun.Load().(bool) {
						pp.proc.Statistics.AddRestart()
						pp.restartWaitGroup.Done()
					} else {
						pp.firstRun.Store(false)
					}
					pp.proc.State = models.Running
				},
				"before_stop": func(e *fsm.Event) {
					isKill := e.Args[0].(bool)
					err := s.stop(name, isKill)
					if err != nil {
						e.Cancel(err)
					}
				},
				"enter_exiting": func(e *fsm.Event) {
					pp.proc.State = models.Exiting
				},
				"before_stopDone": func(e *fsm.Event) {
					pp.proc.Statistics.AddStop()
					pp.proc.Statistics.SetLastStopTime()
					s.release(pp)
				},
				"enter_exited": func(e *fsm.Event) {
					state := e.Args[0].(*os.ProcessState)
					pp.proc.State = models.Exited
					pp.done(state)
				},
				"before_remove": func(e *fsm.Event) {
					s.procMap.Delete(name)
				},
			},
		),
	}

	pp.firstRun.Store(true)
	_, loaded = s.procMap.LoadOrStore(name, pp)
	if loaded {
		return nil, ProcessExist
	}

	err := pp.FSM.Event("init")
	if err != nil {
		return nil, err
	}

	return pp.proc, nil
}

func (s *processesService) FindByName(name string) *models.Process {
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

	err := pp.FSM.Event("start")
	if err != nil {
		return err
	}
	return nil
}

func (s *processesService) Stop(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	err := pp.FSM.Event("stop", false)
	if err != nil {
		return err
	}
	return nil
}

func (s *processesService) Kill(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	err := pp.FSM.Event("stop", true)
	if err != nil {
		return err
	}
	return nil
}

func (s *processesService) Restart(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	err := pp.FSM.Event("stop", false)
	if err != nil {
		return err
	}

	pp.Wait()

	err = pp.FSM.Event("start")
	if err != nil {
		return err
	}
	return nil
}

func (s *processesService) Remove(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	err := pp.FSM.Event("remove", true)
	if err != nil {
		return err
	}
	return nil
}

func (s *processesService) Watch(name string) error {
	result, ok := s.procMap.Load(name)
	if !ok {
		return ProcessNotExist
	}
	pp := result.(*processPair)

	return s.watchDog.StartWatch(pp.proc, pp)
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
		pp.Wait()
		close(ret)
	}()

	return ret
}

func (s *processesService) WaitRestart(name string) <-chan error {
	ret := make(chan error, 1)
	result, ok := s.procMap.Load(name)
	if !ok {
		ret <- ProcessNotExist
		return ret
	}
	pp := result.(*processPair)

	go func() {
		pp.restartWaitGroup.Wait()
		ret <- nil
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
