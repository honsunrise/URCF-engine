package core

import (
	"context"
	"crypto/tls"
	"errors"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

var ErrCannotConnectTwice = errors.New("can't connect twice")

type warpPlugin struct {
	name      string
	rpcClient *rpc.Client
}

func (wp *warpPlugin) Command(name string, params []string) (string, error) {
	rpcCli := wp.rpcClient
	var result string
	err := rpcCli.Call("Plugin.Command", nil, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (wp *warpPlugin) GetHelp(name string) (string, error) {
	rpcCli := wp.rpcClient
	var help string
	err := rpcCli.Call("Plugin.GetHelp", nil, &help)
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
	return &JsonRPCServer{
		rpi: rpi,
	}, nil
}

type JsonRPCServer struct {
	rpcClientMap sync.Map
	httpServer   *http.Server
	rpi          RegisterPluginInterface
}

func (s *JsonRPCServer) getPluginInfo(rpcClient *rpc.Client) error {
	info := &PluginReportInfo{}
	err := rpcClient.Call("Plugin.GetPluginInfo", nil, info)
	if err != nil {
		return err
	}
	_, loaded := s.rpcClientMap.LoadOrStore(info.Name, rpcClient)
	if loaded {
		return ErrCannotConnectTwice
	}
	err = s.rpi.Register(info.Name, &warpPlugin{
		name:      info.Name,
		rpcClient: rpcClient,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *JsonRPCServer) Serve(lis net.Listener, TLS *tls.Config) error {
	http.Handle("/plugin", websocket.Handler(func(ws *websocket.Conn) {
		rpcClient := jsonrpc.NewClient(ws)
		err := s.getPluginInfo(rpcClient)
		if err != nil {
			ws.Close()
		}
	}))
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
