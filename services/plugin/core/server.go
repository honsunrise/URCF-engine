package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/utils"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ServerNotRun = errors.New("server not run")

var CoreProtocolVersion, _ = utils.NewSemVerFromString("1.0.0-rc1")

type Protocol int32

const (
	NoneProtocol Protocol = iota
	GRPCProtocol
)

var protocolStrings = []utils.IntName{
	{0, "NoneProtocol"},
	{1, "GRPCProtocol"},
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
	EnvRequestVersion       = "ENV_REQUEST_VERSION"

	MsgCoreVersion = "CoreVersion"
	MsgVersion     = "Version"
	MsgAddress     = "Address"
	MsgRpcProtocol = "RPCProtocol"
	MsgDone        = "DONE"
)

type ServerConfig struct {
	Address              net.Addr
	SupportProtocols     Protocols
	TLS                  *tls.Config
	ConnectedCallback    func(server *Server)
	DisconnectedCallback func(server *Server)
}

type serverStatus int

const (
	serverStatusTimeOut serverStatus = iota
	serverStatusStopped
	serverStatusPartInit
	serverStatusEarlyExit
	serverStatusDone
)

type Server struct {
	lock     sync.Mutex
	plugins  map[string]serverInstanceInterface
	context  context.Context
	config   *ServerConfig
	process  *models.Process
	protocol Protocol
	status   serverStatus
}

func NewServer(config *ServerConfig) (*Server, error) {
	if config.SupportProtocols == nil {
		config.SupportProtocols = Protocols{
			GRPCProtocol,
		}
	}

	if config.Address == nil {
		addr, err := GetRandomListenerAddr(true)
		if err != nil {
			return nil, err
		}
		config.Address = addr
	}

	return &Server{
		config:  config,
		context: context.Background(),
		status:  serverStatusStopped,
	}, nil
}

//func (c *Server) exitCleanUp() {
//	procServ := processes.GetInstance()
//	<-procServ.Wait(c.config.Name)
//	c.lock.Lock()
//	defer c.lock.Unlock()
//	c.status = serverStatusStopped
//}

func (c *Server) Start() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	env := make(map[string]string)
	env[EnvPluginConnectAddress] = utils.CovertToSchemeAddress(c.config.Address)
	env[EnvSupportRpcProtocol] = c.config.SupportProtocols.String()
	env[EnvRequestVersion] = c.config.Version.String()

	procServ := processes.GetInstance()
	process, err := procServ.Prepare(c.config.Name, c.config.WorkDir, c.config.Cmd, c.config.Args,
		env, models.HookLog|models.AutoRestart)
	if err != nil && err != processes.ProcessExist {
		return err
	} else if err == processes.ProcessExist {
		c.process = procServ.FindByName(c.config.Name)
	} else {
		c.process = process
	}

	err = procServ.Start(c.config.Name)
	if err != nil {
		return err
	}
	go func() {
		<-procServ.Wait(c.config.Name)
		c.config.DisconnectedCallback(c)
	}()
	go func() {
		err := <-procServ.WaitRestart(c.config.Name)
		if err != nil {
			return
		}
		err = c.communication()
		if err != nil {
			return
		}
		c.config.ConnectedCallback(c)
	}()

	return nil
}

func (c *Server) communication() (err error) {
	procServ := processes.GetInstance()
	linesCh := make(chan []byte)
	go func() {
		defer close(linesCh)

		buf := bufio.NewReader(c.process.DataOut)
		for {
			line, err := buf.ReadBytes('\n')
			if err == io.EOF {
				return
			}

			if line != nil {
				linesCh <- line
			}
		}
	}()

	defer func() {
		go func() {
			for range linesCh {
			}
		}()
	}()

	defer func() {
		if c.status != clientStatusDone && c.status != clientStatusEarlyExit {
			c.status = clientStatusStopped
			procServ.Stop(c.config.Name)
		}
	}()

	timeout := time.After(c.config.StartTimeout)
	for true {
		select {
		case <-timeout:
			c.status = clientStatusTimeOut
			return errors.New("timeout while waiting for plugin to start")
		case <-procServ.Wait(c.config.Name):
			c.status = clientStatusEarlyExit
			return errors.New("plugin exited before we could connect")
		case lineBytes := <-linesCh:
			line := strings.TrimSpace(string(lineBytes))
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				err = fmt.Errorf("Unrecognized remote plugin message: %s\n", line)
				return err
			}
			parts[0] = strings.TrimSpace(parts[0])
			parts[1] = strings.TrimSpace(parts[1])
			c.status = clientStatusPartInit
			switch strings.ToLower(parts[0]) {
			case strings.ToLower(MsgCoreVersion):
				var coreProtocol *utils.SemanticVersion
				coreProtocol, err = utils.NewSemVerFromString(parts[1])
				if err != nil {
					return err
				}

				if !CoreProtocolVersion.Compatible(coreProtocol) {
					err = fmt.Errorf("Incompatible core API version with plugin. "+
						"Plugin version: %s, Core version: %s\n\n"+
						"Please report this to the plugin author.", coreProtocol, CoreProtocolVersion)
					return err
				}
			case strings.ToLower(MsgVersion):
				var protocol *utils.SemanticVersion
				protocol, err = utils.NewSemVerFromString(parts[1])
				if err != nil {
					return err
				}

				if !c.config.Version.Compatible(protocol) {
					err = fmt.Errorf("Incompatible API version with plugin. "+
						"Plugin version: %s, Request version: %s", protocol, c.config.Version)
					return err
				}
			case strings.ToLower(MsgAddress):
				addr := utils.ParseSchemeAddress(parts[1])
				if addr == nil {
					err = fmt.Errorf("Unsupported address format: %s", parts[1])
					return err
				}
				c.config.Address = addr
			case strings.ToLower(MsgRpcProtocol):
				ui64 := uint64(0)
				ui64, err = strconv.ParseUint(parts[1], 10, 0)
				if err != nil {
					return err
				}
				c.protocol = Protocol(ui64)
				if !c.config.SupportProtocols.Exist(c.protocol) {
					err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
						c.protocol, c.config.SupportProtocols)
					return err
				}
			case strings.ToLower(MsgDone):
				switch c.protocol {
				case NoneProtocol:
				case GRPCProtocol:
					c.rpcClient, err = NewGRPCClient(c.context, c.config)
					if err != nil {
						return err
					}
				default:
					err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
						c.protocol, c.config.SupportProtocols)
					return err
				}

				err = c.rpcClient.Initialization()
				if err != nil {
					return err
				}
				c.status = clientStatusDone

				go c.exitCleanUp()
				return nil
			}
		}
	}
	return nil
}

func (c *Server) Deploy(name string) (interface{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != clientStatusDone {
		return NoneProtocol, ServerNotRun
	}
	return c.rpcClient.Deploy(name)
}

func (c *Server) Protocol() (Protocol, error) {
	if c.status != clientStatusDone {
		return NoneProtocol, ServerNotRun
	}
	return c.protocol, nil
}

func (c *Server) Stop() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != clientStatusDone {
		return ServerNotRun
	}
	procServ := processes.GetInstance()
	c.rpcClient.UnInitialization()
	err := procServ.Stop(c.config.Name)
	if err != nil {
		return err
	}
	return nil
}
