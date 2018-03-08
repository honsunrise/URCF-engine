package daemon

import (
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"github.com/sevlyar/go-daemon"
	"os"
	"path"
	"syscall"
)

func GetCtx() *daemon.Context {
	confServ := global_configuration.GetGlobalConfig()
	workPath := confServ.Get().Sys.WorkPath
	if _, err := os.Stat(workPath); os.IsNotExist(err) {
		os.MkdirAll(workPath, 0755)
	}
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

func IsDaemonRunning(ctx *daemon.Context) (bool, *os.Process, error) {
	d, err := ctx.Search()

	if err != nil {
		return false, d, err
	}

	if err := d.Signal(syscall.Signal(0)); err != nil {
		return false, d, err
	}

	return true, d, nil
}