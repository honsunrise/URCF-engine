package autostart

import (
	"sort"
	"sync"
	"time"

	"github.com/kataras/iris/core/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/autostart"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/services/processes/types"
)

type Service interface {
	services.ServiceLifeCycle
	StartAll() error
	EnableAll() error
	DisableAll() error
	Add(process types.Process, startDelay int32, stopDelay int32, priority int32, parallel bool) (int64, error)
	Remove(id int64) error
	Disable(id int64) error
	Enable(id int64) error
}

type autoStart struct {
	services.InitHelper
	sync.Mutex
	init             bool
	processes        map[string]*types.Process
	repo             autostart.Repository
	cache            map[int64]*models.AutoStart
	sortCacheKey     []string
	processesService processes.Service
}

func (a *autoStart) Initialize(arguments ...interface{}) error {
	return a.CallInitialize(func() error {
		return nil
	})
}

func (a *autoStart) UnInitialize(arguments ...interface{}) error {
	return a.CallUnInitialize(func() error {
		return nil
	})
}

var instance *autoStart
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &autoStart{
			cache:            make(map[int64]*models.AutoStart),
			sortCacheKey:     make([]string, 0),
			processes:        make(map[string]*types.Process),
			repo:             autostart.NewAutostartRepository(),
			init:             false,
			processesService: processes.GetInstance(),
		}
	})
	return instance
}

func initCache(a *autoStart) error {
	a.Lock()
	defer a.Unlock()
	if !a.init {
		a.init = true
		all, err := a.repo.FindAll()
		if err != nil {
			return err
		}
		sort.Sort(sort.Reverse(models.ByPriority(all)))
		for _, as := range all {
			a.cache[as.ID] = &as
		}
	}
	return nil
}

func (a *autoStart) StartAll() error {
	err := initCache(a)
	if err != nil {
		return err
	}
	go func() {
		for _, as := range a.cache {
			if as.Parallel {
				a.processes[as.Name], err = a.processesService.Prepare(as.Name, as.WorkDir, as.Cmd,
					as.Args, as.Env, as.Option)
				if err != nil {
					log.Errorf("process %s autostart error: %v", as.Name, err)
				}
				err = a.processesService.Start(as.Name)
				if err != nil {
					log.Errorf("process %s autostart error: %v", as.Name, err)
				}
			}
		}
		for _, as := range a.cache {
			if !as.Parallel {
				<-time.After(time.Second * time.Duration(as.StartDelay))
				a.processes[as.Name], err = a.processesService.Prepare(as.Name, as.WorkDir, as.Cmd,
					as.Args, as.Env, as.Option)
				if err != nil {
					log.Warnf("process %s autostart error: %v", as.Name, err)
				}
				err = a.processesService.Start(as.Name)
				if err != nil {
					log.Errorf("process %s autostart error: %v", as.Name, err)
				}
			}
		}
	}()
	return err
}

func (a *autoStart) EnableAll() error {
	err := initCache(a)
	if err != nil {
		return err
	}
	a.Lock()
	defer a.Unlock()
	for k, v := range a.cache {
		a.cache[k].Enable = true
		a.repo.UpdateAutoStartByID(v.ID, map[string]interface{}{
			"Enable": true,
		})
	}
	return err
}

func (a *autoStart) DisableAll() error {
	err := initCache(a)
	if err != nil {
		return err
	}
	a.Lock()
	defer a.Unlock()
	for k, v := range a.cache {
		a.cache[k].Enable = true
		a.repo.UpdateAutoStartByID(v.ID, map[string]interface{}{
			"Enable": false,
		})
	}
	return err
}

func (a *autoStart) Add(process types.Process, startDelay int32, stopDelay int32, priority int32, parallel bool) (id int64, err error) {
	err = initCache(a)
	if err != nil {
		return
	}
	as := &models.AutoStart{
		Priority:     priority,
		StartDelay:   startDelay,
		StopDelay:    stopDelay,
		Enable:       true,
		Parallel:     parallel,
		ProcessParam: process.ProcessParam,
	}
	a.Lock()
	defer a.Unlock()
	err = a.repo.InsertAutoStart(as)
	if err != nil {
		return
	}
	id = as.ID
	return
}

func (a *autoStart) Remove(id int64) error {
	err := initCache(a)
	if err != nil {
		return err
	}
	if a.cache[id] != nil {
		a.Lock()
		defer a.Unlock()
		delete(a.cache, id)
		_, err = a.repo.DeleteAutoStartByID(id)
		if err != nil {
			return err
		}
	}
	return err
}

func (a *autoStart) Disable(id int64) error {
	err := initCache(a)
	if err != nil {
		return err
	}
	as := a.cache[id]
	if as == nil {
		return errors.New("can't find target process")
	}
	a.Lock()
	defer a.Unlock()
	as.Enable = false
	a.repo.UpdateAutoStartByID(id, map[string]interface{}{
		"Enable": false,
	})
	return err
}

func (a *autoStart) Enable(id int64) error {
	err := initCache(a)
	if err != nil {
		return err
	}
	as := a.cache[id]
	if as == nil {
		return errors.New("can't find target process")
	}
	a.Lock()
	defer a.Unlock()
	as.Enable = true
	a.repo.UpdateAutoStartByID(id, map[string]interface{}{
		"Enable": true,
	})
	return err
}
