package plugin_test

import (
	"fmt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"github.com/zhsyourai/URCF-engine/utils/async"
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
	err := pluginService.Start("HelloWord")
	if err != nil {
		t.Fatalf("%s(%s)", "Start error", fmt.Sprint(err))
	}
	<-pluginService.ListCommand("HelloWord").Subscribe(
		async.ErrFunc(func(err error) {
			t.Fatalf("%s(%s)", "List command error", fmt.Sprint(err))
		}), async.ResultFunc(func(result interface{}) {
			t.Logf("List command: %v", result)
		}),
	)
	<-pluginService.Command("HelloWord", "Hello").Subscribe(
		async.ErrFunc(func(err error) {
			t.Fatalf("%s(%s)", "Call Hello command error", fmt.Sprint(err))

		}), async.ResultFunc(func(result interface{}) {
			t.Logf("Hello command result: %v", result)

		}),
	)
	<-time.After(time.Second * 3)
}
