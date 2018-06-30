package core

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/zhsyourai/URCF-engine/utils/async"
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

func (wp *warpPlugin) Command(name string, params []string) async.AsyncRet {
	return async.From(
		func() interface{} {
			rpcCli := wp.rpcClient
			commands := make([]string, 10)
			err := rpcCli.Call("Plugin.Command", nil, &commands)
			if err != nil {
				return err
			}
			return commands
		})
}

func (wp *warpPlugin) GetHelp(name string) async.AsyncRet {
	return async.From(
		func() interface{} {
			rpcCli := wp.rpcClient
			var help string
			err := rpcCli.Call("Plugin.GetHelp", nil, &help)
			if err != nil {
				return err
			}
			return help
		})
}

func (wp *warpPlugin) ListCommand() async.AsyncRet {
	return async.From(
		func() interface{} {
			rpcCli := wp.rpcClient
			commands := make([]string, 10)
			err := rpcCli.Call("Plugin.ListCommand", nil, &commands)
			if err != nil {
				return err
			}
			return commands
		})
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
	return s.httpServer.Serve(lis)
}

func (s *JsonRPCServer) Stop() error {
	return s.httpServer.Shutdown(context.Background())
}
