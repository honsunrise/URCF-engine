package plugin

import (
	"sync"

	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/plugin"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"path"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol"
)

type InstallFlag int32

const (
	Reinstall InstallFlag = 1 << iota
	Purge
)

type UnInstallFlag int32

const (
	KeepConfig UnInstallFlag = 1 << iota
)

type Service interface {
	services.ServiceLifeCycle
	GetAll() ([]models.Plugin, error)
	GetByName(name string) (models.Plugin, error)
	Uninstall(name string, flag UnInstallFlag) error
	Install(path string, flag InstallFlag) (models.Plugin, error)
	GetInterface(name string) protocol.CommandProtocol
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

func (s *pluginService) GetByName(name string) (models.Plugin, error) {
	return models.Plugin{}, nil
}

func (s *pluginService) Uninstall(name string, flag UnInstallFlag) error {
	return nil
}

func (f *pluginService) checkArchitecture() error {
	return nil
}

func (f *pluginService) checkOS() error {
	return nil
}

func (f *pluginService) checkSum() error {
	return nil
}

func (f *pluginService) checkSysDeps() error {
	return nil
}

func (f *pluginService) checkDeps() error {
	return nil
}

func (s *pluginService) Install(filePath string, flag InstallFlag) (plugin models.Plugin, err error) {
	pluginFile, err := Open(filePath)
	if err != nil {
		return
	}

	err = s.checkArchitecture()
	if err != nil {
		return
	}

	err = s.checkOS()
	if err != nil {
		return
	}

	err = s.checkSum()
	if err != nil {
		return
	}

	err = s.checkDeps()
	if err != nil {
		return
	}

	err = s.checkDeps()
	if err != nil {
		return
	}

	confServ := global_configuration.GetGlobalConfig()
	releasePath := path.Join(confServ.Get().Sys.PluginPath, pluginFile.PluginManifest.Name +
		pluginFile.PluginManifest.Version.String())

	pluginFile.ReleaseToDirectory(releasePath)
	if err != nil {
		return
	}

	err = s.repo.InsertPlugin(&plugin)
	if err != nil {
		return
	}

	return
}

func (s *pluginService) GetInterface(name string) protocol.CommandProtocol {
	return nil
}
