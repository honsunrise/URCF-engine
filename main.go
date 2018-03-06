package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/http"
	"github.com/zhsyourai/URCF-engine/rpc"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/sevlyar/go-daemon"
	"github.com/zhsyourai/URCF-engine/services/configuration"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	logService "github.com/zhsyourai/URCF-engine/services/log"
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/zhsyourai/URCF-engine/services/netfilter"
	"github.com/zhsyourai/URCF-engine/services/processes/autostart"
	"github.com/zhsyourai/URCF-engine/services/processes/watchdog"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/zhsyourai/URCF-engine/services/plugin"
)

var (
	app = kingpin.New("urcf-engine", "Universal Remote Config Framework Engine")

	serve      = app.Command("serve", "Create URCF daemon.")
	configFile = serve.Flag("config-file", "Config file location").String()

	serveStop = app.Command("kill", "Kill daemon URCF.")

	version        = app.Command("version", "get version")
	currentVersion = "0.1.0"
)

func main() {
	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	if *configFile == "" {
		folderPath := os.Getenv("HOME") + "/.URCF"
		*configFile = folderPath + "/config.yml"
		os.MkdirAll(folderPath, 0755)
	}
	switch command {
	case serveStop.FullCommand():
		stopServer()
	case serve.FullCommand():
		startServer()
	case version.FullCommand():
		fmt.Println(currentVersion)
	}
}

func start(ctx *daemon.Context) (err error) {
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

func stop(ctx *daemon.Context) (err error) {
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

func isDaemonRunning(ctx *daemon.Context) (bool, *os.Process, error) {
	d, err := ctx.Search()

	if err != nil {
		return false, d, err
	}

	if err := d.Signal(syscall.Signal(0)); err != nil {
		return false, d, err
	}

	return true, d, nil
}

func getCtx() *daemon.Context {
	confServ := global_configuration.GetGlobalConfig()
	workPath := confServ.Get().Sys.WorkPath
	ctx := &daemon.Context{
		PidFileName: path.Join(workPath, "main.pid"),
		PidFilePerm: 0644,
		LogFileName: path.Join(workPath, "main.log"),
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}
	return ctx
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

func startServer() {
	gConfServ := global_configuration.GetGlobalConfig()
	gConfServ.Initialize(*configFile)
	ctx := getCtx()
	if ok, _, _ := isDaemonRunning(ctx); ok {
		log.Info("Server daemon is already running.")
		return
	}

	d, err := ctx.Reborn()
	if err != nil {
		log.Fatalf("Failed to reborn daemon due to %+v.", err)
	}

	if d != nil {
		if waitForStartResult(d) {
			log.Info("Server daemon started")
		} else {
			log.Info("Server daemon start failed, detail see log file")
		}
		return
	}
	defer ctx.Release()

	log.Info("Starting server daemon...")
	err = start(ctx)
	if err != nil {
		log.Fatal(err)
	}
	sendSignal(os.Getppid(), syscall.SIGUSR1)
	log.Info("Server daemon started")
	sigKill := make(chan os.Signal, 1)
	signal.Notify(sigKill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigKill
	log.Info("Received signal to stop...")
	err = stop(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gConfServ.UnInitialize(*configFile)
	os.Exit(0)
}

func stopServer() {
	gConfServ := global_configuration.GetGlobalConfig()
	gConfServ.Initialize(*configFile)
	log.Info("Stopping server daemon ...")
	ctx := getCtx()
	defer ctx.Release()
	if ok, p, err := isDaemonRunning(ctx); ok {
		if err := p.Signal(syscall.Signal(syscall.SIGQUIT)); err != nil {
			log.Fatalf("Failed to kill server daemon %v", err)
		}
	} else {
		if err == nil {
			log.Fatal("Search server install error", err)
		} else {
			log.Info("Server Instance is not running.")
		}
	}
	log.Info("Server daemon terminated")
	gConfServ.UnInitialize(*configFile)
	os.Exit(0)
}
