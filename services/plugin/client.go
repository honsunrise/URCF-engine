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
	"strconv"
	log "github.com/sirupsen/logrus"
	"errors"
)

type ClientConfig struct {
	Plugins          map[string]interface{}
	Version
	Name             string
	Cmd              string
	Args             []string
	WorkDir          string
	MinPort, MaxPort uint
	StartTimeout     time.Duration
}

type Client struct {
	sync.Mutex
	context     context.Context
	config      *ClientConfig
	process     *types.Process
	address     net.Addr
}

func NewClient(config *ClientConfig) *Client {
	return &Client{
		config: config,
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
	case <-procServ.WaitChan(process):
		err = errors.New("plugin exited before we could connect")
	case lineBytes := <-linesCh:
		// Trim the line and split by "|" in order to get the parts of
		// the output.
		line := strings.TrimSpace(string(lineBytes))
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 4 {
			err = fmt.Errorf(
				"Unrecognized remote plugin message: %s\n\n"+
					"This usually means that the plugin is either invalid or simply\n"+
					"needs to be recompiled to support the latest protocol.", line)
			return err
		}

		// Check the core protocol. Wrapped in a {} for scoping.
		{
			var coreProtocol int64
			coreProtocol, err = strconv.ParseInt(parts[0], 10, 0)
			if err != nil {
				err = fmt.Errorf("Error parsing core protocol version: %s", err)
				return err
			}

			if int(coreProtocol) != CoreProtocolVersion {
				err = fmt.Errorf("Incompatible core API version with plugin. "+
					"Plugin version: %s, Core version: %d\n\n"+
					"To fix this, the plugin usually only needs to be recompiled.\n"+
					"Please report this to the plugin author.", parts[0], CoreProtocolVersion)
				return err
			}
		}

		// Parse the protocol version
		var protocol int64
		protocol, err = strconv.ParseInt(parts[1], 10, 0)
		if err != nil {
			err = fmt.Errorf("Error parsing protocol version: %s", err)
			return err
		}

		// Test the API version
		if uint(protocol) != c.config.ProtocolVersion {
			err = fmt.Errorf("Incompatible API version with plugin. "+
				"Plugin version: %s, Core version: %d", parts[1], c.config.ProtocolVersion)
			return err
		}

		switch parts[2] {
		case "tcp":
			c.address, err = net.ResolveTCPAddr("tcp", parts[3])
		case "unix":
			c.address, err = net.ResolveUnixAddr("unix", parts[3])
		default:
			err = fmt.Errorf("Unknown address type: %s", parts[3])
		}

		// If we have a server type, then record that. We default to net/rpc
		// for backwards compatibility.
		c.protocol = ProtocolNetRPC
		if len(parts) >= 5 {
			c.protocol = Protocol(parts[4])
		}

		found := false
		for _, p := range c.config.AllowedProtocols {
			if p == c.protocol {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
				c.protocol, c.config.AllowedProtocols)
			return err
		}
	}

	return err
}

func (c *Client) Stop() error {

	return nil
}
