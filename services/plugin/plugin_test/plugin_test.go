package plugin_test

import (
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol"
	"github.com/zhsyourai/URCF-engine/utils"
	"testing"
	"time"
)

func TestPluginService(t *testing.T) {
	stub := protocol.NewPluginStub()
	client, err := stub.StartUp(&models.Plugin{
		ID:          "test_hello_world",
		Title:       "Hello World!",
		Enabled:     true,
		InstallDate: time.Now(),
		Path:        "./plugin_sdk",
		WorkDir:     "./hello_world",
		EnterPoint: []string{
			"/usr/bin/python3",
			"plugin.py",
		},
		Version: *utils.SemanticVersionMust(utils.NewSemVerFromString("1.0.0")),
	})
	if err != nil {
		t.Fatal(err)
	}
	commands, err := client.ListCommand()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Commands %v", commands)
	if len(commands) != 1 {
		t.Fatal("Command len not correct!")
	}
	if commands[0] != "Hello" {
		t.Fatal("Command hello not supported!")
	}
	result, err := client.Command("Hello")
	if err != nil {
		t.Fatal(err)
	}
	if result != "World" {
		t.Fatal("Command Hello exec not correct")
	}
	t.Logf("Exec result %v", result)
	err = stub.Stop()
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(time.Second * 3)
}
