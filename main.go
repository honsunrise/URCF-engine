package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/http"
	"github.com/zhsyourai/URCF-engine/rpc"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/sevlyar/go-daemon"
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
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case serveStop.FullCommand():
		stopServer()
	case serve.FullCommand():
		startServer()
	case version.FullCommand():
		fmt.Println(currentVersion)
	}
}

func start() (err error) {
	err = rpc.StartRPCServer()
	if err != nil {
		log.Fatal(err)
	}
	err = http.StartHTTPServer()
	if err != nil {
		log.Fatal(err)
	}
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
	if *configFile == "" {
		folderPath := os.Getenv("HOME")
		*configFile = folderPath + "/.URCF/config.yml"
		os.MkdirAll(path.Dir(*configFile), 0755)
	}

	ctx := &daemon.Context{
		PidFileName: path.Join(filepath.Dir(*configFile), "main.pid"),
		PidFilePerm: 0644,
		LogFileName: path.Join(filepath.Dir(*configFile), "main.log"),
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}
	return ctx
}

func waitForStartResult() bool {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1, syscall.SIGUSR2)
	waitedSignal := <-signalChan
	if waitedSignal == syscall.SIGUSR1 {
		return true
	}
	return false
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
		waitForStartResult()
		log.Info("Server daemon started")
		return
	}
	defer ctx.Release()

	log.Info("Starting server daemon...")
	err = start()
	if err != nil {
		log.Fatal(err)
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
	}
	os.Exit(0)
}

func stopServer() {
	log.Info("Stopping server daemon ...")
	ctx := getCtx()
	if ok, p, _ := isDaemonRunning(ctx); ok {
		if err := p.Signal(syscall.Signal(syscall.SIGQUIT)); err != nil {
			log.Fatalf("Failed to kill server daemon %v", err)
		}
	} else {
		ctx.Release()
		log.Info("Server Instance is not running.")
	}
	log.Info("Server daemon terminated")
}
