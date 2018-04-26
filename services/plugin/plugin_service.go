package plugin

import (
	"fmt"
	"github.com/kataras/iris/core/errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/repositories/plugin"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol"
	"io"
	"os"
	"path"
	"strings"
	"sync"
)

const None = 0

type InstallFlag int32

const (
	Reinstall InstallFlag = 1 << iota
)

var (
	ErrPluginNotRun = errors.New("Plugin is not running")
)

func ParseInstallFlag(option string) (ret InstallFlag, err error) {
	ret = None
	options := strings.Split(option, ",")
	for _, opt := range options {
		switch opt {
		case "reinstall":
			ret |= Reinstall
		case "none":
			return None, nil
		default:
			return None, fmt.Errorf("not a valid InstallFlag: %q", opt)
		}
	}
	return ret, nil
}

type UninstallFlag int32

const (
	KeepConfig UninstallFlag = 1 << iota
)

func ParseUninstallFlag(option string) (ret UninstallFlag, err error) {
	ret = None
	options := strings.Split(option, ",")
	for _, opt := range options {
		switch opt {
		case "keepConfig":
			ret |= KeepConfig
		case "none":
			return None, nil
		default:
			return None, fmt.Errorf("not a valid UninstallFlag: %q", opt)
		}
	}
	return ret, nil
}

type Service interface {
	services.ServiceLifeCycle
	ListAll(page uint32, size uint32, sort string, order string) (int64, []models.Plugin, error)
	GetByName(name string) (models.Plugin, error)
	Uninstall(name string, flag UninstallFlag) error
	Install(path string, flag InstallFlag) (models.Plugin, error)
	InstallByReaderAt(readerAt io.ReaderAt, size int64, flag InstallFlag) (models.Plugin, error)
	Start(name string) (protocol.CommandProtocol, error)
	Stop(name string) error
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
	stubMap sync.Map
	repo    plugin.Repository
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

func (s *pluginService) ListAll(page uint32, size uint32, sort string, order string) (total int64, plugins []models.Plugin,
	err error) {
	total, err = s.repo.CountAll()
	if err != nil {
		return 0, []models.Plugin{}, err
	}
	if sort == "" {
		plugins, err = s.repo.FindAll(page, size, nil)
		if err != nil {
			return 0, []models.Plugin{}, err
		}
	} else {
		o, err := repositories.ParseOrder(order)
		if err != nil {
			return 0, []models.Plugin{}, err
		}
		plugins, err = s.repo.FindAll(page, size, []repositories.Sort{
			{
				Name:  sort,
				Order: o,
			},
		})
		if err != nil {
			return 0, []models.Plugin{}, err
		}
	}

	return
}

func (s *pluginService) GetByName(name string) (models.Plugin, error) {
	return s.GetByName(name)
}

func (s *pluginService) Uninstall(name string, flag UninstallFlag) error {
	return nil
}

func (s *pluginService) Install(filePath string, flag InstallFlag) (plugin models.Plugin, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}
	return s.InstallByReaderAt(file, fileInfo.Size(), flag)
}

func (s *pluginService) InstallByReaderAt(readerAt io.ReaderAt, size int64,
	flag InstallFlag) (plugin models.Plugin, err error) {
	pluginFile, err := OpenReader(readerAt, size)
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
	releasePath := path.Join(confServ.Get().Sys.PluginPath, pluginFile.PluginManifest.Name+"@"+
		pluginFile.PluginManifest.Version)

	err = pluginFile.ReleaseToDirectory(releasePath)
	if err != nil {
		return
	}

	err = s.repo.InsertPlugin(&plugin)
	if err != nil {
		return
	}

	return
}

func (s *pluginService) Start(name string) (cp protocol.CommandProtocol, err error) {
	confServ := global_configuration.GetGlobalConfig()
	p, err := s.repo.FindPluginByName(name)
	if err != nil {
		return
	}
	if value, ok := s.stubMap.Load(name); ok {
		stub := value.(*protocol.PluginStub)
		cp, err = stub.GetPluginInterface()
		return
	}
	home := path.Join(confServ.Get().Sys.PluginHome, name)
	if _, err := os.Stat(home); os.IsNotExist(err) {
		os.MkdirAll(home, 0770)
	}
	stub, err := protocol.StartUpPluginStub(&p, home)
	if err != nil {
		return
	}
	cp, err = stub.GetPluginInterface()
	if err != nil {
		return
	}
	s.stubMap.Store(name, stub)
	return
}

func (s *pluginService) Stop(name string) (err error) {
	if value, ok := s.stubMap.Load(name); ok {
		stub := value.(*protocol.PluginStub)
		err = stub.Stop()
		return
	} else {
		return ErrPluginNotRun
	}
	return
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
