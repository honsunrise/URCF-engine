package account

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/zhsyourai/URCF-engine/rpc/client"
	"fmt"
)

func Prepare(app *kingpin.Application) map[string]func() error {
	account := app.Command("account", "account operation")
	rpcAddress := account.Flag("rpc-address", "the urcf serve rpc address").
		Default("localhost:8228").TCP()

	register := account.Command("register", "register account")
	userId := register.Arg("id", "user id").String()
	password := register.Arg("password", "user password").String()
	roles := register.Arg("role", "user role").Strings()

	return map[string]func() error{
		register.FullCommand(): func() error {
			rpc, err := client.NewAccountRPC((*rpcAddress).String())
			if err != nil {
				return err
			}
			account, err := rpc.Register(*userId, *password, *roles)
			if err != nil {
				return err
			}
			fmt.Printf("New Account is %v \n", account)
			return nil
		},
	}
}
