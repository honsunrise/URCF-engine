package plugin

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
)

type Protocol int32

const (
	GRPCProtocol Protocol = 1 << iota
)

type ClientConfig struct {
	Plugins          map[string]interface{}
	Version          SemanticVersion
	Name             string
	Cmd              string
	Args             []string
	WorkDir          string
	MinPort, MaxPort uint
	StartTimeout     time.Duration
	AllowedProtocols Protocol
}

type Client struct {
	sync.Mutex
	context  context.Context
	config   *ClientConfig
	process  *types.Process
	address  net.Addr
	protocol Protocol
}

func NewClient(config *ClientConfig) *Client {
	return &Client{
		config:  config,
		context: context.Background(),
	}
}

func (c *Client) Start() error {
	procServ := processes.GetInstance()
	process, err := procServ.Prepare(c.config.Name, c.config.WorkDir, c.config.Cmd, c.config.Args, nil, models.HookLog)
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
	select {
	case <-timeout:
		err = errors.New("timeout while waiting for plugin to start")
	case <-procServ.WaitExitChan(process):
		err = errors.New("plugin exited before we could connect")
	case lineBytes := <-linesCh:
		line := strings.TrimSpace(string(lineBytes))
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 5 {
			err = fmt.Errorf(
				"Unrecognized remote plugin message: %s\n\n"+
					"This usually means that the plugin is either invalid or simply\n"+
					"needs to be recompiled to support the latest protocol.", line)
			return err
		}

		// Check the core protocol version
		{
			var coreProtocol *SemanticVersion
			coreProtocol, err = NewSemVerFromString(parts[0])
			if err != nil {
				return err
			}

			if CoreProtocolVersion.Compatible(coreProtocol) {
				err = fmt.Errorf("Incompatible core API version with plugin. "+
					"Plugin version: %s, Core version: %s\n\n"+
					"To fix this, the plugin usually only needs to be recompiled.\n"+
					"Please report this to the plugin author.", coreProtocol, CoreProtocolVersion)
				return err
			}
		}
		// Check the API protocol version
		{
			var protocol *SemanticVersion
			protocol, err = NewSemVerFromString(parts[1])
			if err != nil {
				return err
			}

			// Test the API version
			if c.config.Version.Compatible(protocol) {
				err = fmt.Errorf("Incompatible API version with plugin. "+
					"Plugin version: %s, Core version: %s", protocol, c.config.Version)
				return err
			}
		}

		switch parts[2] {
		case "tcp":
			c.address, err = net.ResolveTCPAddr("tcp", parts[3])
		case "unix":
			c.address, err = net.ResolveUnixAddr("unix", parts[3])
		default:
			err = fmt.Errorf("Unknown address type: %s", parts[3])
		}

		ui64 := uint64(0)
		ui64, err = strconv.ParseUint(parts[4], 10, 0)
		if err != nil {
			return err
		}
		c.protocol = Protocol(ui64)

		if (c.config.AllowedProtocols & c.protocol) == 0 {
			err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
				c.protocol, c.config.AllowedProtocols)
			return err
		}


	}
	return nil
}

func (c *Client) Stop() error {

	return nil
}
