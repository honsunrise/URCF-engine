package commands

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/zhsyourai/URCF-engine/commands/version"
	"fmt"
	"github.com/zhsyourai/URCF-engine/commands/serve"
	"github.com/zhsyourai/URCF-engine/commands/kill"
	"github.com/zhsyourai/URCF-engine/commands/account"
	"os"
)

var app = kingpin.New("urcf", "Universal Remote Config Framework Engine")

var registry = make(map[string]func() error)

func register(processors map[string]func() error)  {
	for command, processor := range processors {
		if registry[command] != nil {
			panic(fmt.Errorf("command %q is already registered", command))
		}
		registry[command] = processor
	}
}

func init() {
	register(version.Prepare(app))
	register(serve.Prepare(app))
	register(kill.Prepare(app))
	register(account.Prepare(app))
}

func Run() int {
	selected := kingpin.MustParse(app.Parse(os.Args[1:]))
	if selected == "" {
		app.Usage(os.Args[1:])
		return -1
	}
	processor := registry[selected]
	if processor == nil {
		panic(fmt.Errorf("command %q not found", selected))
	}
	if processor() != nil {
		return -1
	}
	return 0
}