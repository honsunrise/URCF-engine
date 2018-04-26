package plugin_test

import (
	"fmt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"testing"
	"time"
)

func TestInstall(t *testing.T) {
	pluginService := plugin.GetInstance()
	_, err := pluginService.Install("./plugin_test/hello_world.ppk", plugin.None)
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
}

func TestStart(t *testing.T) {
	pluginService := plugin.GetInstance()
	cp, err := pluginService.Start("HelloWord")
	if err != nil {
		t.Fatalf("%s(%s)", "Start error", fmt.Sprint(err))
	}
	commands, err := cp.ListCommand()
	if err != nil {
		t.Fatalf("%s(%s)", "List command error", fmt.Sprint(err))
	}
	t.Logf("List command: %v", commands)
	result, err := cp.Command("Hello")
	if err != nil {
		t.Fatalf("%s(%s)", "Call Hello command error", fmt.Sprint(err))
	}
	t.Logf("Hello command result: %v", result)
	<-time.After(time.Second * 3)
}
