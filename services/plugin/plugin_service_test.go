package plugin_test

import (
	"fmt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"testing"
)

func TestInstall(t *testing.T) {
	pluginService := plugin.GetInstance()
	_, err := pluginService.Install("./plugin_test/hello_world.ppk", plugin.None)
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
}
