package serve

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/zhsyourai/URCF-engine/daemon"
	"os"
	"github.com/zhsyourai/URCF-engine/services/netfilter"
	"github.com/zhsyourai/URCF-engine/services/processes/watchdog"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"syscall"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"os/signal"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/http"
	"github.com/zhsyourai/URCF-engine/rpc"
	"github.com/zhsyourai/URCF-engine/services/configuration"
	logService "github.com/zhsyourai/URCF-engine/services/log"
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/zhsyourai/URCF-engine/services/processes/autostart"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"github.com/kataras/iris/core/errors"
)

func Prepare(app *kingpin.Application) (*kingpin.CmdClause, func() error) {
	serve      := app.Command("serve", "Create URCF daemon.")
	configFile := serve.Flag("config-file", "Config file location").String()
	startAsDaemon := serve.Flag("daemon", "Config file location").Default("false").Bool()

	return serve, func() error {
		if *configFile == "" {
			folderPath := os.Getenv("HOME") + "/.URCF"
			*configFile = folderPath + "/config.yml"
			os.MkdirAll(folderPath, 0755)
		}
		*startAsDaemon = true

		gConfServ := global_configuration.GetGlobalConfig()
		gConfServ.Initialize(*configFile)
		defer gConfServ.UnInitialize(*configFile)
		if *startAsDaemon {
			ctx := daemon.GetCtx()
			defer ctx.Release()

			if ok, _, _ := daemon.IsDaemonRunning(ctx); ok {
				return errors.New("server daemon is already running.")
			}

			d, err := ctx.Reborn()
			if err != nil {
				return err
			}

			if d != nil {
				if waitForStartResult(d) {
					log.Info("Server daemon started")
				} else {
					return errors.New("Server daemon start failed, detail see log file")
				}
				return nil
			}
			log.Info("Starting server daemon...")
			return run()
		} else {
			return run()
		}
		return nil
	}
}

func run() error {
	err := start()
	if err != nil {
		log.Fatal(err)
		return err
	}
	sendSignal(os.Getppid(), syscall.SIGUSR1)
	log.Info("Server daemon started")
	sigKill := make(chan os.Signal, 1)
	signal.Notify(sigKill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigKill
	log.Info("Received signal to stop...")
	err = stop()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}


func start() (err error) {
	confServ := configuration.GetInstance()
	confServ.Initialize()
	accountServ := account.GetInstance()
	accountServ.Initialize()
	logServ := logService.GetInstance()
	logServ.Initialize()
	netfilterServ := netfilter.GetInstance()
	netfilterServ.Initialize()
	watchdogServ := watchdog.GetInstance()
	watchdogServ.Initialize()
	autostartServ := autostart.GetInstance()
	autostartServ.Initialize()
	processesServ := processes.GetInstance()
	processesServ.Initialize()
	pluginServ := plugin.GetInstance()
	pluginServ.Initialize()
	go func() {
		err = rpc.StartRPCServer()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err = http.StartHTTPServer()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return
}

func stop() (err error) {
	err = rpc.StopRPCServer()
	if err != nil {
		log.Fatal(err)
	}
	err = http.StopHTTPServer()
	if err != nil {
		log.Fatal(err)
	}
	pluginServ := plugin.GetInstance()
	pluginServ.UnInitialize()
	processesServ := processes.GetInstance()
	processesServ.UnInitialize()
	autostartServ := autostart.GetInstance()
	autostartServ.UnInitialize()
	watchdogServ := watchdog.GetInstance()
	watchdogServ.UnInitialize()
	netfilterServ := netfilter.GetInstance()
	netfilterServ.UnInitialize()
	logServ := logService.GetInstance()
	logServ.UnInitialize()
	accountServ := account.GetInstance()
	accountServ.UnInitialize()
	confServ := configuration.GetInstance()
	confServ.UnInitialize()
	return
}

func waitForStartResult(p *os.Process) bool {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1, syscall.SIGUSR2)
	ok := make(chan bool)
	go func() {
		waitedSignal := <-signalChan
		if waitedSignal == syscall.SIGUSR1 {
			ok <- true
		}
		ok <- false
	}()

	go func() {
		p.Wait()
		ok <- false
	}()
	return <-ok
}

func sendSignal(pid int, signal os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	defer p.Release()
	return p.Signal(signal)
}

