package version

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

func Prepare(app *kingpin.Application) map[string]func() error {
	version := app.Command("version", "get version")
	currentVersion := "0.1.0"

	return map[string]func() error{
		version.FullCommand(): func() error {
			fmt.Println(currentVersion)
			return nil
		},
	}
}
