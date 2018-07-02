package core

import (
	"crypto/tls"
	"net"
)

type PluginInterface interface {
	Command(name string, params []string) (string, error)
	GetHelp(name string) (string, error)
	ListCommand() ([]string, error)
}

type ServerInterface interface {
	Serve(lis net.Listener, TLS *tls.Config) error
	Stop() error
}

type RegisterPluginInterface interface {
	Register(name string, plugin PluginInterface) error
}

type ServerFactory interface {
	New(rpi RegisterPluginInterface) (ServerInterface, error)
}
