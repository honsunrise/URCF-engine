package plugin_test

import (
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol"
	"github.com/zhsyourai/URCF-engine/utils"
	"testing"
	"time"
)

func TestPluginService(t *testing.T) {
	stub, err := protocol.StartUpPluginStub(&models.Plugin{
		Name:        "test_hello_world",
		Enable:      true,
		InstallDir:  "./hello_world",
		InstallTime: time.Now(),
		EnterPoint:  "/usr/bin/python3 plugin.py",
		Version:     *utils.SemanticVersionMust(utils.NewSemVerFromString("1.0.0")),
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := stub.GetPluginInterface()
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
