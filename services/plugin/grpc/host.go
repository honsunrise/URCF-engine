package grpc

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"time"
	"net"
	"sync"
	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
	"crypto/tls"
	"google.golang.org/grpc/credentials"
)

type Host struct {
	quit chan struct{}
	quitOnce sync.Once
	serveWG  sync.WaitGroup
}

func NewHost() *Host {
	return &Host{
		quit: make(chan struct{}),
	}
}

func (h *Host) Serve() error {
	// Register a listener so we can accept a connection
	listener, err := plugin.Listener()
	if err != nil {
		log.Error("plugin host start error", err)
		return err
	}

	h.serveWG.Add(1)
	defer func() {
		listener.Close()
		h.serveWG.Done()
		select {
		case <-h.quit:
			<-s.done
		}
	}()

	var tempDelay time.Duration

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(*net.OpError); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				select {
				case <-time.After(tempDelay):
					continue
				case <-h.quit:
					return nil
				}
			}

			select {
			case <-h.quit:
				return nil
			default:
			}
			return err
		}
		tempDelay = 0

		h.serveWG.Add(1)
		go func() {
			h.handleConn(conn)
			h.serveWG.Done()
		}()
	}

	return nil
}

func (h *Host) Stop() error {
	h.quitOnce.Do(func() {
		close(h.quit)
	})
	return nil
}

func (h *Host) handleConn(connect net.Conn) {
	session, err := yamux.Client(connect, yamux.DefaultConfig())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	// Build dialing options.
	opts := make([]grpc.DialOption, 0, 5)
	opts = append(opts, grpc.WithDialer(func(s string, duration time.Duration) (net.Conn, error) {
		return session.Open()
	}))
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.FailOnNonTempDialError(true))
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("unused", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	go func() {
		NewPluginInterfaceClient(conn)
	}()
}