package watchdog

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/processes/types"
)

type Service interface {
	services.ServiceLifeCycle
	StartWatch(proc *types.Process) error
	StartWatchWithNotify(proc *types.Process, notify <-chan *os.ProcessState) error
	StopWatch(proc *types.Process) error
	GetDeathsChan() chan *types.Process
}

type dog struct {
	Stopping   atomic.Value
	StopNotify chan struct{}
	ExitNotify chan *os.ProcessState
	Proc       *types.Process
}

type watchDog struct {
	services.InitHelper
	sync.Mutex
	deathProcesses chan *types.Process
	watchProcesses map[string]*dog
}

func (watcher *watchDog) Initialize(arguments ...interface{}) error {
	return watcher.CallInitialize(func() error {
		return nil
	})
}

func (watcher *watchDog) UnInitialize(arguments ...interface{}) error {
	return watcher.CallUnInitialize(func() error {
		return nil
	})
}

var instance *watchDog
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &watchDog{
			deathProcesses: make(chan *types.Process),
			watchProcesses: make(map[string]*dog),
		}
	})
	return instance
}

func waitTargetProcess(proc *types.Process, dog *dog) {
	log.Infof("Starting watcher on process %s", proc.Name)
	state, _ := proc.Process.Wait()
	if dog.Stopping.Load().(bool) {
		return
	}
	dog.Stopping.Store(true)
	dog.ExitNotify <- state
}

func waitTargetNotify(proc *types.Process, notify *sync.WaitGroup, dog *dog) {
	log.Infof("Starting watcher on process %s with notify", proc.Name)
	state := notify.Wait
	if dog.Stopping.Load().(bool) {
		return
	}
	dog.Stopping.Store(true)
	dog.ExitNotify <- state
}

func (watcher *watchDog) watch(proc *types.Process, dog *dog) {
	defer delete(watcher.watchProcesses, proc.Name)
	select {
	case procStatus := <-dog.ExitNotify:
		log.Infof("Proc %s is dead, advising master...", proc.Name)
		log.Infof("State is %s", procStatus.String())
		watcher.deathProcesses <- proc
		break
	case <-dog.StopNotify:
		break
	}
}

func (watcher *watchDog) StartWatch(proc *types.Process) (err error) {
	watcher.Lock()
	defer watcher.Unlock()
	if _, ok := watcher.watchProcesses[proc.Name]; ok {
		log.Warnf("A watcher for this process already exists.")
		return
	}
	dog := &dog{
		Proc:       proc,
		ExitNotify: make(chan *os.ProcessState, 1),
		StopNotify: make(chan struct{}, 1),
	}
	dog.Stopping.Store(false)
	watcher.watchProcesses[proc.Name] = dog
	go waitTargetProcess(proc, dog)
	go watcher.watch(proc, dog)
	return
}

func (watcher *watchDog) StartWatchWithNotify(proc *types.Process, notify *sync.WaitGroup) (err error) {
	watcher.Lock()
	defer watcher.Unlock()
	if _, ok := watcher.watchProcesses[proc.Name]; ok {
		log.Warnf("A watcher for this process already exists.")
		return
	}
	dog := &dog{
		Proc:       proc,
		ExitNotify: make(chan *os.ProcessState, 1),
		StopNotify: make(chan struct{}, 1),
	}
	dog.Stopping.Store(false)
	watcher.watchProcesses[proc.Name] = dog
	go waitTargetNotify(proc, notify, dog)
	go watcher.watch(proc, dog)
	return
}

func (watcher *watchDog) StopWatch(proc *types.Process) error {
	watcher.Lock()
	defer watcher.Unlock()
	if dog, ok := watcher.watchProcesses[proc.Name]; ok {
		log.Infof("Exiting watcher on proc %s", proc.Name)
		if dog.Stopping.Load().(bool) {
			return errors.New("watch is stopping")
		}
		dog.Stopping.Store(true)
		dog.StopNotify <- struct{}{}
		return nil
	}
	return errors.New("process not watching")
}

func (watcher *watchDog) GetDeathsChan() chan *types.Process {
	return watcher.deathProcesses
}
