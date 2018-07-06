package plugin

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/core/errors"
	"github.com/looplab/fsm"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/repositories/plugin"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/utils"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"
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
	ErrPluginNotStarted     = errors.New("Plugin is not started")
	ErrPluginIsStarting     = errors.New("Plugin is starting")
	ErrPluginExist          = errors.New("plugin exist")
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
	Command(pluginName string, name string, params ...string) (string, error)
	GetHelp(pluginName string, name string) (string, error)
	ListCommand(pluginName string) ([]string, error)
}

var instance *pluginService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &pluginService{
			repo: plugin.NewPluginRepository(),
			supportProtocols: map[models.Protocol]core.ServerFactory{
				models.JsonRPCProtocol: &core.JsonRPCFactory{},
			},
			protocols: make(map[models.Protocol]core.ServerInterface),
		}
	})
	return instance
}

type pluginPair struct {
	FSM             *fsm.FSM
	PluginInterface core.PluginInterface
	process         *models.Process
}

type pluginService struct {
	services.InitHelper
	repo             plugin.Repository
	supportProtocols map[models.Protocol]core.ServerFactory
	protocols        map[models.Protocol]core.ServerInterface
	plugins          sync.Map
	config           global_configuration.ServerConfig
}

func (s *pluginService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		s.config = global_configuration.GetGlobalConfig().Get().PluginServer
		if s.config.UsedProtocols == nil || len(s.config.Address) == 0 {
			s.config.UsedProtocols = models.Protocols{
				models.JsonRPCProtocol,
			}
		}

		if s.config.Address == nil || len(s.config.Address) == 0 {
			s.config.Address = make(map[models.Protocol]net.Listener)
			for _, protocol := range s.config.UsedProtocols {
				addr, err := Listener(true)
				if err != nil {
					return err
				}
				s.config.Address[protocol] = addr
			}
		}

		if s.config.TLS == nil {
			s.config.TLS = make(map[models.Protocol]*tls.Config)
		}

		for _, protocol := range s.config.UsedProtocols {
			factory := s.supportProtocols[protocol]
			if factory != nil {
				instance, err := s.supportProtocols[protocol].New(s)
				if err != nil {
					return err
				}
				if s.config.TLS[protocol] != nil {
					err := instance.Serve(s.config.Address[protocol], s.config.TLS[protocol])
					if err != nil {
						return err
					}
				} else {
					err := instance.Serve(s.config.Address[protocol], nil)
					if err != nil {
						return err
					}
				}
				s.protocols[protocol] = instance
			} else {
				panic(errors.New("protocol not support"))
			}
		}
		return nil
	})
}

func (s *pluginService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		var err error = nil
		for _, protocol := range s.config.UsedProtocols {
			err1 := s.protocols[protocol].Stop()
			if err1 != nil {
				err = err1
			}
			delete(s.protocols, protocol)
		}
		s.plugins.Range(func(key, value interface{}) bool {
			s.plugins.Delete(key)
			return true
		})
		return err
	})
}

func (s *pluginService) getListenAddress() map[models.Protocol]string {
	ret := make(map[models.Protocol]string, 10)
	for k, v := range s.config.Address {
		ret[k] = utils.CovertToSchemeAddress(v.Addr())
	}
	return ret
}

func (s *pluginService) Register(name string, plugin core.PluginInterface) error {
	if result, ok := s.plugins.Load(name); ok {
		pp := result.(*pluginPair)
		err := pp.FSM.Event("startDone", plugin)
		if err != nil {
			return err
		}
		return nil
	} else {
		return ErrPluginNotStarted
	}
}

func (s *pluginService) UnRegister(name string) error {
	if result, ok := s.plugins.Load(name); ok {
		pp := result.(*pluginPair)
		err := pp.FSM.Event("stopDone")
		if err != nil {
			return err
		}
		return nil
	} else {
		return ErrPluginNotStarted
	}
}

func (s *pluginService) _fsmHelper(name string) (core.PluginInterface, error) {
	if v, ok := s.plugins.Load(name); ok {
		pp := v.(*pluginPair)
		if pp.FSM.Is("started") {
			return pp.PluginInterface, nil
		} else if pp.FSM.Is("starting") {
			return nil, ErrPluginIsStarting
		}
		return nil, ErrPluginNotStarted
	} else {
		return nil, ErrPluginNotStarted
	}
}

func (s *pluginService) Command(name string, command string, params ...string) (string, error) {
	if pluginInterface, err := s._fsmHelper(name); err != nil {
		return "", err
	} else {
		result, err := pluginInterface.Command(command, params)
		if err != nil {
			return "", err
		}
		return result, nil
	}
}

func (s *pluginService) GetHelp(name string, command string) (string, error) {
	if pluginInterface, err := s._fsmHelper(name); err != nil {
		return "", err
	} else {
		result, err := pluginInterface.GetHelp(command)
		if err != nil {
			return "", err
		}
		return result, nil
	}
}

func (s *pluginService) ListCommand(name string) ([]string, error) {
	if pluginInterface, err := s._fsmHelper(name); err != nil {
		return nil, err
	} else {
		result, err := pluginInterface.ListCommand()
		if err != nil {
			return nil, err
		}
		return result, nil
	}
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

func (s *pluginService) start(name string) error {
	procServ := processes.GetInstance()
	err := procServ.Start(name)
	if err != nil {
		return err
	}
	go func() {
		<-procServ.Wait(name)
		result, _ := s.plugins.Load(name)
		pp := result.(*pluginPair)
		pp.FSM.Event("stopDone")
	}()
	return nil
}

func (s *pluginService) stop(name string) error {
	go func() {
		time.Sleep(10 * time.Second)
		result, _ := s.plugins.Load(name)
		pp := result.(*pluginPair)
		pp.FSM.Event("stopDone")
	}()
	procServ := processes.GetInstance()
	err := procServ.Stop(name)
	if err != nil {
		return err
	}
	return nil
}

func (s *pluginService) stopDone(name string) error {
	s.plugins.Delete(name)
	return nil
}

func (s *pluginService) Start(name string) error {
	config := global_configuration.GetGlobalConfig().Get().PluginServer
	p, err := s.repo.FindPluginByName(name)
	if err != nil {
		return err
	}
	_, loaded := s.plugins.Load(p.Name)
	if loaded {
		return ErrPluginHasBeenStarted
	}
	listenAddr := s.getListenAddress()
	jsonListenAddr, err := json.Marshal(listenAddr)
	if err != nil {
		return err
	}
	enterPoint := strings.Split(p.EnterPoint, " ")
	env := make(map[string]string)
	env[EnvPluginConnectAddress] = string(jsonListenAddr)
	env[EnvSupportRpcProtocol] = config.UsedProtocols.String()
	env[EnvInstallVersion] = p.Version.String()

	procServ := processes.GetInstance()
	process, err := procServ.Prepare(p.Name, p.InstallDir, enterPoint[0], enterPoint[1:], env,
		models.HookLog|models.AutoRestart)
	if err != nil && err != processes.ProcessExist {
		return err
	} else if err == processes.ProcessExist {
		process = procServ.FindByName(p.Name)
	}
	pp := &pluginPair{
		process: process,
	}
	pp.FSM = fsm.NewFSM("stopped",
		fsm.Events{
			{Name: "start", Src: []string{"stopped", "starting"}, Dst: "starting"},
			{Name: "startDone", Src: []string{"starting", "started"}, Dst: "started"},
			{Name: "stop", Src: []string{"started", "stopping"}, Dst: "stopping"},
			{Name: "stopDone", Src: []string{"started", "stopping", "stopped"}, Dst: "stopped"},
		},
		fsm.Callbacks{
			"leave_stopped": func(e *fsm.Event) {
				err := s.start(name)
				if err != nil {
					e.Cancel(err)
				}
			},
			"leave_starting": func(e *fsm.Event) {
				pi := e.Args[0].(core.PluginInterface)
				pp.PluginInterface = pi
			},
			"leave_started": func(e *fsm.Event) {
				if e.Dst == "stopping" {
					err := s.stop(name)
					if err != nil {
						e.Cancel(err)
					}
				} else {
					err := s.stopDone(name)
					if err != nil {
						e.Cancel(err)
					}
				}
			},
			"leave_stopping": func(e *fsm.Event) {
				err := s.stopDone(name)
				if err != nil {
					e.Cancel(err)
				}
			},
		},
	)
	_, loaded = s.plugins.LoadOrStore(p.Name, pp)
	if loaded {
		return ErrPluginHasBeenStarted
	}

	err = pp.FSM.Event("start")
	if err != nil {
		return err
	}
	return nil
}

func (s *pluginService) Stop(name string) error {
	result, loaded := s.plugins.Load(name)
	if loaded {
		pp := result.(*pluginPair)
		err := pp.FSM.Event("stop")
		if err != nil {
			return err
		}
		return nil
	} else {
		return ErrPluginNotStarted
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
