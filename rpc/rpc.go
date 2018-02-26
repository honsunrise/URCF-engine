package rpc

import (
	"net/rpc"
	"net/http"
	"net"
	"github.com/zhsyourai/URCF-engine/services/configuration"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/services/account"
)

func StartRPCServer() (err error) {
	confServ := configuration.GetInstance()
	address := confServ.Get("system.rpc.address").Value.(string)
	err = rpc.RegisterName("AccountRPC", &AccountRPC{
		service: account.GetInstance(),
	})
	if err != nil {
		log.Fatal("Register Account RPC error:", err)
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
	return
}

func StopRPCServer() (err error) {
	return
}
