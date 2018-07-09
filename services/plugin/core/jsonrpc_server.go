package core

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/prometheus/common/log"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"sync"
	"time"
)

var ErrCannotConnectTwice = errors.New("can't connect twice")

type warpPlugin struct {
	name      string
	rpcClient *jsonrpc2.Client
	exit      chan struct{}
}

type commandParam struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

func (wp *warpPlugin) Command(name string, params []string) (string, error) {
	rpcCli := wp.rpcClient
	var result string
	err := rpcCli.Call("Plugin.Command", &commandParam{
		Name:   name,
		Params: params,
	}, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (wp *warpPlugin) GetHelp(name string) (string, error) {
	rpcCli := wp.rpcClient
	var help string
	err := rpcCli.Call("Plugin.GetHelp", name, &help)
	if err != nil {
		return "", err
	}
	return help, nil

}

func (wp *warpPlugin) ListCommand() ([]string, error) {
	rpcCli := wp.rpcClient
	commands := make([]string, 10)
	err := rpcCli.Call("Plugin.ListCommand", nil, &commands)
	if err != nil {
		return nil, err
	}
	return commands, nil
}

type JsonRPCFactory struct{}

func (JsonRPCFactory) New(rpi RegisterPluginInterface) (ServerInterface, error) {
	server := &JsonRPCServer{
		rpi: rpi,
	}
	return server, nil
}

type JsonRPCServer struct {
	httpServer *http.Server
	rpi        RegisterPluginInterface
	connMap    sync.Map
}

func (s *JsonRPCServer) serverPlugin(ws *websocket.Conn) error {
	rpcClient := jsonrpc2.NewClient(ws)
	info := &PluginReportInfo{}
	var err error = nil
	err = rpcClient.Call("Plugin.GetPluginInfo", nil, info)
	if err != nil {
		return err
	}
	result, loaded := s.connMap.Load(info.Name)
	if loaded {
		lastWp := result.(*warpPlugin)
		var pong string
		err := lastWp.rpcClient.Call("Plugin.Ping", nil, &pong)
		if err != nil {
			s.rpi.UnRegister(info.Name)
			s.connMap.Delete(info.Name)
			close(lastWp.exit)
		} else {
			return ErrCannotConnectTwice
		}
	}
	wp := &warpPlugin{
		name:      info.Name,
		rpcClient: rpcClient,
		exit:      make(chan struct{}),
	}
	err = s.rpi.Register(info.Name, wp)
	if err != nil {
		return err
	}

	s.connMap.Store(info.Name, wp)

	for true {
		select {
		case <-time.After(10 * time.Second):
			var pong string
			err := rpcClient.Call("Plugin.Ping", nil, &pong)
			if err != nil {
				return err
			}
		case <-wp.exit:
			return nil
		}
	}
	return nil
}

func (s *JsonRPCServer) Serve(lis net.Listener, TLS *tls.Config) error {
	handler := websocket.Server{Handler: func(ws *websocket.Conn) {
		err := s.serverPlugin(ws)
		if err != nil {
			log.Error(err)
		}
	}}
	http.Handle("/plugin", handler)
	if TLS != nil {
		lis = tls.NewListener(lis, TLS)
	}
	s.httpServer = &http.Server{Handler: nil}
	go s.httpServer.Serve(lis)
	return nil
}

func (s *JsonRPCServer) Stop() error {
	return s.httpServer.Shutdown(context.Background())
}
