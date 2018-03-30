package protocol

import (
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol/grpc"
	"github.com/kataras/iris/core/errors"
)

type PluginStub struct {
	coreClient *core.Client

}

func NewPluginStub() *PluginStub {
	return &PluginStub{
	}
}

func (p *PluginStub) StartUp(plugin *models.Plugin) error {
	coreClient, err := core.NewClient(&core.ClientConfig{
		Plugins: map[string]core.PluginInterface{
			"command": &grpc.CommandPlugin{},
		},
		Version: &plugin.Version,
		Name: plugin.ID,
		Cmd: plugin.EnterPoint[0],
		Args: plugin.EnterPoint[1:],
		WorkDir: plugin.WorkDir,
	})
	if err != nil {
		return err
	}
	p.coreClient = coreClient

	err =coreClient.Start()
	if err != nil {
		return err
	}

	tmpClient, err := coreClient.Deploy("command")
	if err != nil {
		return err
	}

	realClient, ok := tmpClient.(grpc.CommandInterfaceClient)
	if !ok {
		return errors.New("Client must be grpc.CommandInterfaceClient")
	}


	return nil
}

