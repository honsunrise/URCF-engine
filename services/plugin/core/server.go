package core

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/zhsyourai/URCF-engine/utils"
	"net"
	"sync"
)

var PluginExist = errors.New("plugin exist")
var PluginNotExist = errors.New("plugin not exist")

type Protocol int32

const (
	NoneProtocol Protocol = iota
	JsonRPCProtocol
)

var protocolStrings = []utils.IntName{
	{0, "NoneProtocol"},
	{1, "JsonRPCProtocol"},
}

func (i Protocol) String() string {
	return utils.StringName(uint32(i), protocolStrings, "plugin.", false)
}
func (i Protocol) GoString() string {
	return utils.StringName(uint32(i), protocolStrings, "plugin.", true)
}

type Protocols []Protocol

func (ps Protocols) String() string {
	var ret string
	for i, p := range ps {
		if i == 0 {
			ret = p.String()
		} else {
			ret += "," + p.String()
		}
	}
	return ret
}

func (ps Protocols) Exist(item Protocol) bool {
	for _, p := range ps {
		if p == item {
			return true
		}
	}
	return false
}

type PluginReportInfo struct {
	Name    string
	Version utils.SemanticVersion
}

const (
	EnvPluginConnectAddress = "ENV_PLUGIN_CONNECT_ADDRESS"
	EnvSupportRpcProtocol   = "ENV_SUPPORT_RPC_PROTOCOL"
	EnvInstallVersion       = "ENV_INSTALLED_VERSION"
)

type ServerConfig struct {
	Address       map[Protocol]net.Listener
	TLS           map[Protocol]*tls.Config
	UsedProtocols Protocols
}

var DefaultServerConfig = &ServerConfig{}

type Server struct {
	lock             sync.Mutex
	supportProtocols map[Protocol]ServerFactory
	protocols        map[Protocol]ServerInterface
	plugins          map[string]PluginInterface
	context          context.Context
	config           *ServerConfig
}

func (s *Server) Register(name string, plugin PluginInterface) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.plugins[name] != nil {
		return PluginExist
	}

	s.plugins[name] = plugin

	return nil
}

func NewServer(config *ServerConfig) (*Server, error) {
	if config.UsedProtocols == nil || len(config.Address) == 0 {
		config.UsedProtocols = Protocols{
			JsonRPCProtocol,
		}
	}

	if config.Address == nil || len(config.Address) == 0 {
		config.Address = make(map[Protocol]net.Listener)
		for _, protocol := range config.UsedProtocols {
			addr, err := Listener(true)
			if err != nil {
				return nil, err
			}
			config.Address[protocol] = addr
		}
	}

	if config.TLS == nil {
		config.TLS = make(map[Protocol]*tls.Config)
	}

	return &Server{
		config:  config,
		context: context.Background(),
		supportProtocols: map[Protocol]ServerFactory{
			JsonRPCProtocol: &JsonRPCFactory{},
		},
	}, nil
}

func (s *Server) Start() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, protocol := range s.config.UsedProtocols {
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
	}

	return nil
}

func (s *Server) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var err error = nil
	for _, protocol := range s.config.UsedProtocols {
		err1 := s.protocols[protocol].Stop()
		if err1 != nil {
			err = err1
		}
		delete(s.protocols, protocol)
	}

	for k, _ := range s.plugins {
		delete(s.plugins, k)
	}

	return err
}

func (s *Server) GetPlugin(pluginName string) (PluginInterface, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.plugins[pluginName] == nil {
		return nil, PluginNotExist
	}
	return s.plugins[pluginName], nil
}

func (s *Server) GetListenAddress() map[Protocol]string {
	ret := make(map[Protocol]string, 10)
	for k, v := range s.config.Address {
		ret[k] = utils.CovertToSchemeAddress(v.Addr())
	}
	return ret
}

func (s *Server) GetUsedProtocol() Protocols {
	return s.config.UsedProtocols
}
