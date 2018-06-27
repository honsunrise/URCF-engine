package protocol

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kataras/iris/core/errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol/grpc"
	"strings"
	"sync/atomic"
)

type PluginStub struct {
	context    context.Context
	coreClient *core.Client
	realClient atomic.Value
}

func StartUpPluginStub(plugin *models.Plugin) (*PluginStub, error) {
	ret := &PluginStub{
		context: context.Background(),
	}
	enterPoint := strings.Split(plugin.EnterPoint, " ")
	coreClient, err := core.NewClient(&core.ClientConfig{
		Plugins: map[string]core.ClientInstanceInterface{
			"command": &grpc.CommandPlugin{},
		},
		Version: &plugin.Version,
		Name:    plugin.Name,
		Cmd:     enterPoint[0],
		Args:    enterPoint[1:],
		WorkDir: plugin.InstallDir,
		ConnectedCallback: func(client *core.Client) {
			tmpClient, err := client.Deploy("command")
			if err != nil {
				return
			}

			protocol, err := client.Protocol()
			if err != nil {
				return
			}

			switch protocol {
			case core.GRPCProtocol:
				realClient, ok := tmpClient.(grpc.CommandInterfaceClient)
				if !ok {
					err = errors.New("Instance must be grpc.CommandInterfaceClient")
					return
				}
				ret.realClient.Store(realClient)
				return
			default:
				err = errors.New("Unsupported protocol")
				return
			}
		},
		DisconnectedCallback: func(client *core.Client) {
			ret.realClient.Store(nil)
		},
	})
	if err != nil {
		return nil, err
	}
	ret.coreClient = coreClient

	err = coreClient.Start()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *PluginStub) Command(name string, params ...string) (string, error) {
	realClient := p.realClient.Load().(grpc.CommandInterfaceClient)
	if realClient != nil {
		commandResp, err := realClient.Command(p.context, &grpc.CommandRequest{
			Name:   name,
			Params: params,
		})
		if err != nil {
			return "", err
		}
		if commandResp.GetOptionalErr() != nil {
			return "", errors.New(commandResp.GetError())
		}
		return commandResp.GetResult(), nil
	}
	return "", errors.New("client not(lose) connection")
}

func (p *PluginStub) GetHelp(name string) (string, error) {
	realClient := p.realClient.Load().(grpc.CommandInterfaceClient)
	if realClient != nil {
		chResp, err := realClient.GetHelp(p.context, &grpc.CommandHelpRequest{
			Subcommand: name,
		})
		if err != nil {
			return "", err
		}
		if chResp.GetOptionalErr() != nil {
			return "", errors.New(chResp.GetError())
		}
		return chResp.GetHelp(), nil
	}
	return "", errors.New("client not(lose) connection")
}

func (p *PluginStub) ListCommand() ([]string, error) {
	realClient := p.realClient.Load().(grpc.CommandInterfaceClient)
	if realClient != nil {
		lcResp, err := realClient.ListCommand(p.context, &empty.Empty{})
		if err != nil {
			return nil, err
		}
		if lcResp.GetOptionalErr() != nil {
			return nil, errors.New(lcResp.GetError())
		}
		return lcResp.GetCommands(), nil
	}
	return nil, errors.New("client not(lose) connection")
}

func (p *PluginStub) Stop() error {
	return p.coreClient.Stop()
}
