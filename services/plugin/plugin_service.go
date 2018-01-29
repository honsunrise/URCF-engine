package plugin

import (
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/plugin"
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
	GetAll() ([]models.Plugin, error)
	GetByID(id string) (models.Plugin, error)
	Uninstall(id string, flag ...UnInstallFlag) error
	Install(path string, flag ...InstallFlag) (models.Plugin, error)
	GetInterface(id string) (error)
}

func NewPluginService(repo plugin.Repository) Service {
	return &pluginService{repo: repo}
}

type pluginService struct {
	repo plugin.Repository
}

func (s *pluginService) GetAll() ([]models.Plugin, error) {
	return nil, nil
}

func (s *pluginService) GetByID(id string) (models.Plugin, error) {
	return models.Plugin{}, nil
}

func (s *pluginService) Uninstall(id string, flag ...UnInstallFlag) error {
	return nil
}

func (s *pluginService) Install(path string, flag ...InstallFlag) (models.Plugin, error) {
	return models.Plugin{}, nil
}

func (s *pluginService) GetInterface(id string) (error) {
	return nil
}
