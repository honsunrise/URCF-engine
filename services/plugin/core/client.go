package core

import (
	"time"
	"sync"
	"net"
	"context"
	"github.com/zhsyourai/URCF-engine/services/processes/types"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/models"
	"io"
	"bufio"
	"strings"
	"fmt"
	log "github.com/sirupsen/logrus"
	"errors"
	"strconv"
	"crypto/tls"
	"github.com/zhsyourai/URCF-engine/utils"
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

func (i Protocol) String() string   { return utils.StringName(uint32(i), protocolStrings, "plugin.", false) }
func (i Protocol) GoString() string { return utils.StringName(uint32(i), protocolStrings, "plugin.", true) }

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
	ENV_PLUGIN_LISTENER_ADDRESS   = "ENV_PLUGIN_LISTENER_ADDRESS"
	ENV_ALLOW_PLUGIN_RPC_PROTOCOL = "ENV_ALLOW_PLUGIN_RPC_PROTOCOL"
	ENV_REQUEST_VERSION           = "ENV_REQUEST_VERSION"

	MSG_COREVERSION  = "CoreVersion"
	MSG_VERSION      = "Version"
	MSG_ADDRESS      = "Address"
	MSG_RPC_PROTOCOL = "RPCProtocol"
	MSG_DONE         = "DONE"
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

type Client struct {
	client   ClientInterface
	lock     sync.Mutex
	context  context.Context
	config   *ClientConfig
	process  *types.Process
	protocol Protocol
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
		addr, err := GetRandomListenerAddr()
		if err != nil {
			return nil, err
		}
		config.Address = addr
	}

	return &Client{
		config:  config,
		context: context.Background(),
	}, nil
}

func (c *Client) Start() error {
	env := make(map[string]string)
	env[ENV_PLUGIN_LISTENER_ADDRESS] = c.config.Address.String()
	env[ENV_ALLOW_PLUGIN_RPC_PROTOCOL] = c.config.AllowedProtocols.String()
	env[ENV_REQUEST_VERSION] = c.config.Version.String()

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
			if line != nil {
				linesCh <- line
			}

			if err == io.EOF {
				return
			}
		}
	}()

	defer func() {
		go func() {
			for _ = range linesCh {
			}
		}()
	}()
	process, err = procServ.Start(process)
	if err != nil {
		return err
	}
	// Some channels for the next step
	timeout := time.After(c.config.StartTimeout)

	// Start looking for the address
	log.Debug("waiting for RPC address", "path", c.config.Cmd)
	for true {
		select {
		case <-timeout:
			return errors.New("timeout while waiting for plugin to start")
		case <-procServ.WaitExitChan(process):
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

			switch strings.ToLower(parts[0]) {
			case strings.ToLower(MSG_COREVERSION):
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
			case strings.ToLower(MSG_VERSION):
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
			case strings.ToLower(MSG_ADDRESS):
				addr := utils.ParseSchemeAddress(parts[1])
				if addr == nil {
					err = fmt.Errorf("Unsupported address format: %s", parts[1])
					return err
				}
			case strings.ToLower(MSG_RPC_PROTOCOL):
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
			case strings.ToLower(MSG_DONE):
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
					return nil
				}
			}
		}
	}
	return nil
}

func (c *Client) Deploy(name string) (interface{}, error) {
	return c.client.Deploy(name)
}

func (c *Client) Protocol() Protocol {
	return c.protocol
}

func (c *Client) Stop() error {
	err := c.client.UnInitialization()
	if err != nil {
		return nil
	}
	return nil
}
