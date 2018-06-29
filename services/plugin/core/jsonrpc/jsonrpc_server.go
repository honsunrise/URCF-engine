package jsonrpc

import (
	"crypto/tls"
	"errors"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
	"github.com/zhsyourai/URCF-engine/utils/async"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

var ErrCannotConnectTwice = errors.New("can't connect twice")
var ErrPluginNotStarted = errors.New("plugin not started")

type JsonRPCServer struct {
	config       *core.ServerConfig
	rpcClientMap sync.Map
}

func NewJsonRPCServer(config *core.ServerConfig) (*JsonRPCServer, error) {
	return &JsonRPCServer{
		config: config,
	}, nil
}

func (s *JsonRPCServer) getPluginInfo(rpcClient *rpc.Client) error {
	info := &core.PluginReportInfo{}
	err := rpcClient.Call("Plugin.GetPluginInfo", nil, info)
	if err != nil {
		return err
	}
	_, loaded := s.rpcClientMap.LoadOrStore(info.Name, rpcClient)
	if loaded {
		return ErrCannotConnectTwice
	}
	return nil
}

func (s *JsonRPCServer) Serve(lis net.Listener) error {
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		rpcClient := jsonrpc.NewClient(ws)
		s.getPluginInfo(rpcClient)
	}))
	if s.config.TLS != nil {
		lis = tls.NewListener(lis, s.config.TLS)
	}
	return http.Serve(lis, nil)
}

func (s *JsonRPCServer) Command(pluginName string, name string, params []string) async.AsyncRet {
	return async.From(
		func() interface{} {
			return nil
		})
}
