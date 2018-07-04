package plugin_test

import (
	"bufio"
	"encoding/json"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/utils"
	"io"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPluginService(t *testing.T) {
	pluginVersion := utils.SemanticVersionMust(utils.NewSemVerFromString("1.0.0"))
	pluginName := "HelloWorld"
	procServ := processes.GetInstance()
	var err error
	server, err := core.NewServer(core.DefaultServerConfig)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		<-time.After(time.Second * 3)
		<-procServ.Wait(pluginName)
		err = server.Stop()
		if err != nil {
			t.Fatal(err)
			return
		}
	}()

	listenAddr := server.GetListenAddress()
	jsonListenAddr, err := json.Marshal(listenAddr)
	if err != nil {
		t.Fatal(err)
		return
	}
	env := make(map[string]string)
	env[plugin.EnvPluginConnectAddress] = string(jsonListenAddr)
	env[plugin.EnvSupportRpcProtocol] = server.GetUsedProtocol().String()
	env[plugin.EnvInstallVersion] = pluginVersion.String()

	process, err := procServ.Prepare(pluginName, "./hello_world",
		"/usr/bin/python3", []string{"plugin.py"}, env, models.None)
	if err != nil && err != processes.ProcessExist {
		t.Fatal(err)
	}
	go func() {
		buf := bufio.NewReader(process.StdErr)
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				if err == io.EOF || err == syscall.EIO || strings.Contains(err.Error(), "file already closed") {
					return
				} else {
					t.Error(err)
				}
			}
			t.Log(strings.Trim(line, "\n"))
		}
	}()
	go func() {
		buf := bufio.NewReader(process.StdOut)
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				if err == io.EOF || err == syscall.EIO || strings.Contains(err.Error(), "file already closed") {
					return
				} else {
					t.Error(err)
				}
			}
			t.Log(strings.Trim(line, "\n"))
		}
	}()

	err = procServ.Start(pluginName)
	if err != nil {
		t.Fatal(err)
	}

	<-procServ.WaitRestart(pluginName)

	pluginInterface, err := server.GetPlugin(pluginName)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := pluginInterface.ListCommand()
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
	result, err := pluginInterface.Command("Hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "World" {
		t.Fatal("Command Hello exec not correct")
	}
	t.Logf("Exec result %v", result)
}
