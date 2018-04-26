package protocol

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kataras/iris/core/errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol/grpc"
	"strings"
)

type PluginStub struct {
	coreClient      *core.Client
	commandProtocol CommandProtocol
}

type warpGrpcCommandProtocolClient struct {
	client  grpc.CommandInterfaceClient
	context context.Context
}

func (wg *warpGrpcCommandProtocolClient) Command(name string, params ...string) (string, error) {
	commandResp, err := wg.client.Command(wg.context, &grpc.CommandRequest{
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

func (wg *warpGrpcCommandProtocolClient) GetHelp(name string) (string, error) {
	chResp, err := wg.client.GetHelp(wg.context, &grpc.CommandHelpRequest{
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

func (wg *warpGrpcCommandProtocolClient) ListCommand() ([]string, error) {
	lcResp, err := wg.client.ListCommand(wg.context, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	if lcResp.GetOptionalErr() != nil {
		return nil, errors.New(lcResp.GetError())
	}
	return lcResp.GetCommands(), nil
}

func StartUpPluginStub(plugin *models.Plugin) (*PluginStub, error) {
	ret := &PluginStub{}
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

func (p *PluginStub) GetPluginInterface() (CommandProtocol, error) {
	if p.commandProtocol != nil {
		return p.commandProtocol, nil
	} else {
		tmpClient, err := p.coreClient.Deploy("command")
		if err != nil {
			return nil, err
		}

		protocol, err := p.coreClient.Protocol()
		if err != nil {
			return nil, err
		}

		switch protocol {
		case core.GRPCProtocol:
			realClient, ok := tmpClient.(grpc.CommandInterfaceClient)
			if !ok {
				return nil, errors.New("Instance must be grpc.CommandInterfaceClient")
			}
			p.commandProtocol = &warpGrpcCommandProtocolClient{
				context: context.Background(),
				client:  realClient,
			}
			return p.commandProtocol, nil
		default:
			return nil, errors.New("Unsupported protocol")
		}
	}
}

func (p *PluginStub) Stop() error {
	return p.coreClient.Stop()
}
