package watchdog

import (
	"sync"
	log "github.com/sirupsen/logrus"
	"os"
	"errors"
	"sync/atomic"
	"github.com/zhsyourai/URCF-engine/services/processes"
)

type WatchDog interface {
	StartWatch(proc *processes.Process) error
	StopWatch(proc *processes.Process) error
	GetDeathsChan() chan *processes.Process
}

type dog struct {
	Stopping   atomic.Value
	StopNotify chan struct{}
	ExitNotify chan *os.ProcessState
	Proc       *processes.Process
}

type watchDog struct {
	sync.Mutex
	deathProcesses chan *processes.Process
	watchProcesses map[string]*dog
}

// NewWatcherDog will create a watchDog instance.
// Returns a watchDog instance.
func NewWatcherDog() WatchDog {
	watcher := &watchDog{
		deathProcesses: make(chan *processes.Process),
		watchProcesses: make(map[string]*dog),
	}
	return watcher
}

func waitTargetProcess(proc *processes.Process, dog *dog) {
	log.Infof("Starting watcher on process %s", proc.Name)
	state, _ := proc.Process.Wait()
	if dog.Stopping.Load().(bool) {
		return
	}
	dog.Stopping.Store(true)
	dog.ExitNotify <- state
}

func (watcher *watchDog) watch(proc *processes.Process, dog *dog) {
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

func (watcher *watchDog) StartWatch(proc *processes.Process) (err error) {
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

func (watcher *watchDog) StopWatch(proc *processes.Process) error {
	watcher.Lock()
	defer watcher.Unlock()
	if dog, ok := watcher.watchProcesses[proc.Name]; ok {
		log.Infof("Stopping watcher on proc %s", proc.Name)
		if dog.Stopping.Load().(bool) {
			return errors.New("watch is stopping")
		}
		dog.Stopping.Store(true)
		dog.StopNotify <- struct{}{}
		return nil
	}
	return errors.New("process not watching")
}

func (watcher *watchDog) GetDeathsChan() chan *processes.Process {
	return watcher.deathProcesses
}
