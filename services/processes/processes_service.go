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
	"io"
	"bufio"
	"unicode"
	"github.com/hashicorp/go-hclog"
	logservice "github.com/zhsyourai/URCF-engine/services/log"
)

type Service interface {
	services.ServiceLifeCycle
	Prepare(name string, workDir string, cmd string, args []string, env map[string]string,
		option models.ProcessOption) (*types.Process, error)
	FindByName(name string) *types.Process
	Start(p *types.Process) (*types.Process, error)
	Stop(p *types.Process) error
	Restart(p *types.Process) (*types.Process, error)
	Kill(p *types.Process) error
	Clean(p *types.Process) error
	Watch(p *types.Process) error
	IsAlive(p *types.Process) bool
	WaitChan(p *types.Process) chan struct{}
}

type processesPair struct {
	proc *types.Process
	procAttr *os.ProcAttr
	finalArgs []string
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
		err = s.hookLog(name, lStdErr)
		if err != nil {
			return
		}
		err = s.hookLog(name, lStdOut)
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
	procAtr := &os.ProcAttr{
		Dir: workDir,
		Env: buildEnv(finalEnv),
		Files: []*os.File{
			rStdIn,
			rStdOut,
			rStdErr,
			rDataOut,
		},
	}

	s.procMap.Store(name, &processesPair{
		proc: proc,
		procAttr: procAtr,
		finalArgs: append([]string{cmd}, args...),
	})

	return
}

func (s *processesService) FindByName(name string) *types.Process {
	if p, ok := s.procMap.Load(name); ok {
		return p.(*processesPair).proc
	}
	return nil
}

func (s *processesService) Start(proc *types.Process) (*types.Process, error) {
	p, ok := s.procMap.Load(proc.Name)
	if !ok || p.(*processesPair).proc == nil{
		return nil, errors.New("process does not exist")
	}

	process, err := os.StartProcess(proc.Cmd, p.(*processesPair).finalArgs, p.(*processesPair).procAttr)
	if err != nil {
		return nil, err
	}

	if proc.Option&models.AutoRestart != 0 {
		err = s.watchDog.StartWatch(proc)
		if err != nil {
			return nil, err
		}
	}

	proc.Process = process
	proc.Pid = proc.Process.Pid
	proc.Statistics.InitUpTime()
	return proc, nil
}

func (s *processesService) Stop(proc *types.Process) error {
	p, ok := s.procMap.Load(proc.Name)
	if !ok || p.(*processesPair).proc == nil{
		return errors.New("process does not exist")
	}
	defer proc.Process.Release()
	err := proc.Process.Signal(syscall.SIGTERM)
	return err
}

func (s *processesService) Restart(p *types.Process) (*types.Process, error) {
	if s.IsAlive(p) {
		err := s.Stop(p)
		if err != nil {
			return nil, err
		}
	}
	process, err := s.Prepare(p.Name, p.WorkDir, p.Cmd, p.Args, p.Env, p.Option)
	if err != nil {
		return nil, err
	}
	return s.Start(process)
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

func (s *processesService) WaitChan(p *types.Process) chan struct{} {
	exitCh := make(chan struct{})
	go func() {
		p.Process.Wait()
		// Mark that we exited
		close(exitCh)
	}()
	return exitCh
}


func (s *processesService) hookLog(name string, r io.Reader) error {
	logServ := logservice.GetInstance()
	logger, err := logServ.GetLogger(name)
	if err != nil {
		return err
	}
	go func() {
		bufR := bufio.NewReader(r)
		for {
			line, err := bufR.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				entry, err := parseJSON(line)
				if err != nil {
					logger.Debug(line)
				} else {
					out := flattenKVPairs(entry.KVPairs)

					logger = logger.WithField("timestamp", entry.Timestamp.Format(hclog.TimeFormat))
					switch hclog.LevelFromString(entry.Level) {
					case log.DebugLevel:
						logger.Debug(entry.Message, out...)
					case log.InfoLevel:
						logger.Info(entry.Message, out...)
					case log.WarnLevel:
						logger.Warn(entry.Message, out...)
					case log.ErrorLevel:
						logger.Error(entry.Message, out...)
					case log.FatalLevel:
						logger.Error(entry.Message, out...)
					}
				}
			}

			if err == io.EOF {
				break
			}
		}
	}();
	return nil
}