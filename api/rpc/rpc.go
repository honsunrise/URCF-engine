package rpc

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/api/rpc/server"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

func StartRPCServer() (err error) {
	confServ := global_configuration.GetGlobalConfig()
	value := confServ.Get()
	address := "localhost:" + strconv.FormatInt(int64(value.Rpc.Port), 10)
	err = server.RegisterAccountRPC()
	if err != nil {
		log.Fatal("Register Account RPC error:", err)
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	log.Info("RPC listen at: ", address)
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func StopRPCServer() (err error) {
	return
}
