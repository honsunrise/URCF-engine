package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/services/plugin/core/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"time"
)

type GRPCClient struct {
	conn    *grpc.ClientConn
	config  *ClientConfig
	context context.Context
	client  proto.PluginInterfaceClient
}

func dialWithAddrAndTls(context context.Context, addr net.Addr, tls *tls.Config) (*grpc.ClientConn, error) {
	// Build dialing options.
	opts := make([]grpc.DialOption, 0, 5)

	// We use a custom dialer so that we can connect over unix domain sockets
	opts = append(opts, grpc.WithDialer(func(_ string, timeout time.Duration) (net.Conn, error) {
		// Connect to the rpcClient
		conn, err := net.DialTimeout(addr.Network(), addr.String(), timeout)
		if err != nil {
			return nil, err
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			// Make sure to set keep alive so that the connection doesn't die
			tcpConn.SetKeepAlive(true)
		}

		return conn, nil
	}))

	// go-plugin expects to block the connection
	opts = append(opts, grpc.WithBlock())

	// Fail right away
	opts = append(opts, grpc.FailOnNonTempDialError(true))

	// If we have no TLS configuration set, we need to explicitly tell grpc
	// that we're connecting with an insecure connection.
	if tls == nil {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(tls)))
	}

	// Connect. Note the first parameter is unused because we use a custom
	// dialer that has the state to see the address.
	conn, err := grpc.DialContext(context, "unused", opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewGRPCClient(context context.Context, config *ClientConfig) (ClientInterface, error) {
	conn, err := dialWithAddrAndTls(context, config.Address, config.TLS)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		config:  config,
		client:  proto.NewPluginInterfaceClient(conn),
		context: context,
		conn:    conn,
	}, nil
}

func (c *GRPCClient) Initialization() error {
	retErr, err := c.client.Initialization(c.context, &proto.Empty{})
	if err != nil {
		return err
	}
	if retErr.GetOptionalErr() != nil {
		return errors.New(retErr.GetError())
	}
	return nil
}

func (c *GRPCClient) UnInitialization() error {
	retErr, err := c.client.UnInitialization(c.context, &proto.Empty{})
	if err != nil {
		return err
	}
	if retErr.GetOptionalErr() != nil {
		return errors.New(retErr.GetError())
	}
	return nil
}

func (c *GRPCClient) Deploy(name string) (interface{}, error) {
	retErr, err := c.client.Deploy(c.context, &proto.DeployRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	if retErr.GetOptionalErr() != nil {
		return nil, errors.New(retErr.GetError())
	}
	p, ok := c.config.Plugins[name]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type: %s", name)
	}
	return p.Instance(c.context, c.conn)
}

func (c *GRPCClient) Ping(name string) error {
	client := grpc_health_v1.NewHealthClient(c.conn)
	_, err := client.Check(c.context, &grpc_health_v1.HealthCheckRequest{
		Service: name,
	})
	return err
}
