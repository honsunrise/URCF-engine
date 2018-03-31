package protocol

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kataras/iris/core/errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol/grpc"
)

type PluginStub struct {
	coreClient *core.Client
}

func NewPluginStub() *PluginStub {
	return &PluginStub{}
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

func (p *PluginStub) StartUp(plugin *models.Plugin) (CommandProtocol, error) {
	coreClient, err := core.NewClient(&core.ClientConfig{
		Plugins: map[string]core.PluginInterface{
			"command": &grpc.CommandPlugin{},
		},
		Version: &plugin.Version,
		Name:    plugin.ID,
		Cmd:     plugin.EnterPoint[0],
		Args:    plugin.EnterPoint[1:],
		WorkDir: plugin.WorkDir,
	})
	if err != nil {
		return nil, err
	}
	p.coreClient = coreClient

	err = coreClient.Start()
	if err != nil {
		return nil, err
	}

	tmpClient, err := coreClient.Deploy("command")
	if err != nil {
		return nil, err
	}

	switch coreClient.Protocol() {
	case core.GRPCProtocol:
		realClient, ok := tmpClient.(grpc.CommandInterfaceClient)
		if !ok {
			return nil, errors.New("Client must be grpc.CommandInterfaceClient")
		}
		return &warpGrpcCommandProtocolClient{
			context: context.Background(),
			client:  realClient,
		}, nil
	default:
		return nil, errors.New("Unsupported protocol")
	}
}

func (p *PluginStub) Stop(name string) error {
	return nil
}
