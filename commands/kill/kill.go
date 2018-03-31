package kill

import (
	"github.com/kataras/iris/core/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/daemon"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"syscall"
)

func Prepare(app *kingpin.Application) map[string]func() error {
	serveStop := app.Command("kill", "Kill daemon URCF.")
	configFile := serveStop.Flag("config-file", "Config file location").String()

	return map[string]func() error{
		serveStop.FullCommand(): func() error {
			if *configFile == "" {
				folderPath := os.Getenv("HOME") + "/.URCF"
				*configFile = folderPath + "/config.yml"
				os.MkdirAll(folderPath, 0755)
			}
			gConfServ := global_configuration.GetGlobalConfig()
			gConfServ.Initialize(*configFile)
			defer gConfServ.UnInitialize(*configFile)
			log.Info("Stopping server daemon ...")
			ctx := daemon.GetCtx()
			defer ctx.Release()
			if ok, p, err := daemon.IsDaemonRunning(ctx); ok {
				if err := p.Signal(syscall.Signal(syscall.SIGQUIT)); err != nil {
					return err
				}
			} else {
				if err == nil {
					return errors.New("Search server instance error")
				} else {
					log.Info("Server Instance is not running.")
				}
			}
			log.Info("Server daemon terminated")
			return nil
		},
	}
}
