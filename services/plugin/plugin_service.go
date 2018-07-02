package plugin

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/core/errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/repositories/plugin"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/utils"
	"github.com/zhsyourai/URCF-engine/utils/async"
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

const (
	EnvPluginConnectAddress = "ENV_PLUGIN_CONNECT_ADDRESS"
	EnvSupportRpcProtocol   = "ENV_SUPPORT_RPC_PROTOCOL"
	EnvInstallVersion       = "ENV_INSTALLED_VERSION"
)

var (
	ErrPluginHasBeenStarted = errors.New("Plugin has been started")
	ErrPluginNotRun         = errors.New("Plugin is not running")
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
	Start(name string) error
	Stop(name string) error
	Command(pluginName string, name string, params ...string) async.AsyncRet
	GetHelp(pluginName string, name string) async.AsyncRet
	ListCommand(pluginName string) async.AsyncRet
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
	server     *core.Server
	processMap sync.Map
	repo       plugin.Repository
}

func (s *pluginService) Command(pluginName string, name string, params ...string) async.AsyncRet {
	return async.From(func() interface{} {
		pluginInterface, err := s.server.GetPlugin(pluginName)
		if err != nil {
			return err
		}
		result, err := pluginInterface.Command(name, params)
		if err != nil {
			return err
		}
		return result
	})
}

func (s *pluginService) GetHelp(pluginName string, name string) async.AsyncRet {
	return async.From(func() interface{} {
		pluginInterface, err := s.server.GetPlugin(pluginName)
		if err != nil {
			return err
		}
		result, err := pluginInterface.GetHelp(name)
		if err != nil {
			return err
		}
		return result
	})
}

func (s *pluginService) ListCommand(pluginName string) async.AsyncRet {
	return async.From(func() interface{} {
		pluginInterface, err := s.server.GetPlugin(pluginName)
		if err != nil {
			return err
		}
		result, err := pluginInterface.ListCommand()
		if err != nil {
			return err
		}
		return result
	})
}

func (s *pluginService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		var err error
		s.server, err = core.NewServer(core.DefaultServerConfig)
		if err != nil {
			return err
		}
		err = s.server.Start()
		if err != nil {
			return err
		}
		return err
	})
}

func (s *pluginService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		var err error
		err = s.server.Stop()
		if err != nil {
			return err
		}
		return err
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
	return s.repo.FindPluginByName(name)
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

	plugin.Name = pluginFile.PluginManifest.Name
	plugin.Desc = pluginFile.PluginManifest.Desc
	plugin.Maintainer = pluginFile.PluginManifest.Maintainer
	plugin.Homepage = pluginFile.PluginManifest.Homepage
	plugin.Version = *utils.SemanticVersionMust(utils.NewSemVerFromString(pluginFile.PluginManifest.Version))
	plugin.Enable = true
	plugin.InstallDir = releasePath
	plugin.WebsDir = pluginFile.PluginManifest.WebsDir
	plugin.CoverFile = pluginFile.PluginManifest.CoverFile
	plugin.EnterPoint = pluginFile.PluginManifest.EnterPoint

	err = s.repo.InsertPlugin(&plugin)
	if err != nil {
		return
	}

	return
}

func (s *pluginService) Start(name string) error {
	p, err := s.repo.FindPluginByName(name)
	if err != nil {
		return err
	}
	_, loaded := s.processMap.Load(p.Name)
	if loaded {
		return ErrPluginHasBeenStarted
	}
	listenAddr := s.server.GetListenAddress()
	jsonListenAddr, err := json.Marshal(listenAddr)
	if err != nil {
		return err
	}
	enterPoint := strings.Split(p.EnterPoint, " ")
	env := make(map[string]string)
	env[EnvPluginConnectAddress] = string(jsonListenAddr)
	env[EnvSupportRpcProtocol] = s.server.GetUsedProtocol().String()
	env[EnvInstallVersion] = p.Version.String()

	procServ := processes.GetInstance()
	process, err := procServ.Prepare(p.Name, p.InstallDir, enterPoint[0], enterPoint[1:], env,
		models.HookLog|models.AutoRestart)
	if err != nil && err != processes.ProcessExist {
		return err
	} else if err == processes.ProcessExist {
		process = procServ.FindByName(p.Name)
	}
	_, loaded = s.processMap.LoadOrStore(p.Name, process)
	if loaded {
		return ErrPluginHasBeenStarted
	}

	err = procServ.Start(p.Name)
	if err != nil {
		return err
	}

	return nil
}

func (s *pluginService) Stop(name string) error {
	_, loaded := s.processMap.Load(name)
	if loaded {
		procServ := processes.GetInstance()
		err := procServ.Stop(name)
		if err != nil {
			return err
		}
	} else {
		return ErrPluginNotRun
	}
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
