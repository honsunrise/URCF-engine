package plugin

import (
	"sync"

	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/plugin"
	"github.com/zhsyourai/URCF-engine/services"
)

type InstallFlag int32

const (
	Reinstall InstallFlag = 1 << iota
)

type UnInstallFlag int32

const (
	KeepConfig UnInstallFlag = 1 << iota
)

type Service interface {
	services.ServiceLifeCycle
	GetAll() ([]models.Plugin, error)
	GetByID(id string) (models.Plugin, error)
	Uninstall(id string, flag UnInstallFlag) error
	Install(path string, flag InstallFlag) (models.Plugin, error)
	GetInterface(id string) (error)
}

var instance *pluginService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &pluginService{
			repo: plugin.NewPluginRepository(),
		}
	})
	return instance
}

type pluginService struct {
	services.InitHelper
	repo plugin.Repository
}

func (s *pluginService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *pluginService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

func (s *pluginService) GetAll() ([]models.Plugin, error) {
	return nil, nil
}

func (s *pluginService) GetByID(id string) (models.Plugin, error) {
	return models.Plugin{}, nil
}

func (s *pluginService) Uninstall(id string, flag UnInstallFlag) error {
	return nil
}

func (s *pluginService) Install(path string, flag InstallFlag) (models.Plugin, error) {
	return models.Plugin{}, nil
}

func (s *pluginService) GetInterface(id string) (error) {
	return nil
}
