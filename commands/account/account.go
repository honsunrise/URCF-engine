package account

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"fmt"
)

func Prepare(app *kingpin.Application) (*kingpin.CmdClause, func() error) {
	version := app.Command("account", "get version")
	currentVersion := "0.1.0"

	return version, func() error {
		fmt.Println(currentVersion)
		return nil
	}
}