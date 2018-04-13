package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/services/processes/types"
	"github.com/zhsyourai/URCF-engine/utils"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

const (
	EnvPluginListenerAddress  = "ENV_PLUGIN_LISTENER_ADDRESS"
	EnvAllowPluginRpcProtocol = "ENV_ALLOW_PLUGIN_RPC_PROTOCOL"
	EnvRequestVersion         = "ENV_REQUEST_VERSION"

	MsgCoreVersion = "CoreVersion"
	MsgVersion     = "Version"
	MsgAddress     = "Address"
	MsgRpcProtocol = "RPCProtocol"
	MsgDone        = "DONE"
)

type ClientConfig struct {
	Plugins          map[string]PluginInterface
	Version          *utils.SemanticVersion
	Name             string
	Cmd              string
	Args             []string
	WorkDir          string
	Address          net.Addr
	StartTimeout     time.Duration
	AllowedProtocols Protocols
	TLS              *tls.Config
}

type clientStatus int

const (
	clientStatusTimeOut clientStatus = iota
	clientStatusStopped
	clientStatusPartInit
	clientStatusEarlyExit
	clientStatusDone
)

type Client struct {
	client   ClientInterface
	lock     sync.Mutex
	context  context.Context
	config   *ClientConfig
	process  *types.Process
	protocol Protocol
	status   clientStatus
}

func NewClient(config *ClientConfig) (*Client, error) {
	if config.Cmd == "" {
		return nil, errors.New("Cmd can't be empty.")
	}

	if config.Plugins == nil {
		return nil, errors.New("Plugins can't be nil. It's must have last one plugin.")
	}

	if config.WorkDir == "" {
		config.WorkDir = "."
	}

	if config.StartTimeout == 0 {
		config.StartTimeout = 1 * time.Minute
	}

	if config.AllowedProtocols == nil {
		config.AllowedProtocols = Protocols{
			GRPCProtocol,
		}
	}

	if config.Version == nil {
		config.Version, _ = utils.NewSemVerFromString("1.0.0")
	}

	if config.Address == nil {
		addr, err := GetRandomListenerAddr(true)
		if err != nil {
			return nil, err
		}
		config.Address = addr
	}

	return &Client{
		config:  config,
		context: context.Background(),
		status:  clientStatusStopped,
	}, nil
}

func (c *Client) exitCleanUp() {
	<-c.process.ExitChan
	c.lock.Lock()
	defer c.lock.Unlock()
	c.status = clientStatusStopped
}

func (c *Client) Start() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	env := make(map[string]string)
	env[EnvPluginListenerAddress] = utils.CovertToSchemeAddress(c.config.Address)
	env[EnvAllowPluginRpcProtocol] = c.config.AllowedProtocols.String()
	env[EnvRequestVersion] = c.config.Version.String()

	procServ := processes.GetInstance()
	process, err := procServ.Prepare(c.config.Name, c.config.WorkDir, c.config.Cmd, c.config.Args, env, models.HookLog)
	if err != nil {
		return err
	}
	c.process = process

	linesCh := make(chan []byte)
	go func() {
		defer close(linesCh)

		buf := bufio.NewReader(process.DataOut)
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
			for _ = range linesCh {
			}
		}()
	}()
	err = procServ.Start(process)
	if err != nil {
		return err
	}

	defer func() {
		if c.status != clientStatusDone && c.status != clientStatusEarlyExit {
			c.status = clientStatusStopped
			procServ.Stop(c.process)
		}
	}()

	timeout := time.After(c.config.StartTimeout)
	for true {
		select {
		case <-timeout:
			c.status = clientStatusTimeOut
			return errors.New("timeout while waiting for plugin to start")
		case <-process.ExitChan:
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
				if !c.config.AllowedProtocols.Exist(c.protocol) {
					err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
						c.protocol, c.config.AllowedProtocols)
					return err
				}
			case strings.ToLower(MsgDone):
				switch c.protocol {
				case NoneProtocol:
				case GRPCProtocol:
					c.client, err = NewGRPCClient(c.context, c.config)
					if err != nil {
						return err
					}
				default:
					err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
						c.protocol, c.config.AllowedProtocols)
					return err
				}

				err = c.client.Initialization()
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

func (c *Client) Deploy(name string) (interface{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.client.Deploy(name)
}

func (c *Client) Protocol() (Protocol, error) {
	if c.status != clientStatusDone {
		return NoneProtocol, errors.New("client not run")
	}
	return c.protocol, nil
}

func (c *Client) Stop() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	procServ := processes.GetInstance()
	if c.status != clientStatusDone {
		return errors.New("client not run")
	}
	c.client.UnInitialization()
	err := procServ.Stop(c.process)
	if err != nil {
		return err
	}
	return nil
}
