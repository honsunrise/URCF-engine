package grpc

import (
	"google.golang.org/grpc"
	"context"
	"net"
	"crypto/tls"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/services/plugin/core"
)

type GRPCClient struct {
	conn    *grpc.ClientConn
	config  *core.ClientConfig
	context context.Context
	client  PluginInterfaceClient
}

func convertNetAddrToGRPCAddr(addr net.Addr) string {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		if len(tcpAddr.IP) == net.IPv4len {
			return "ipv4://" + tcpAddr.String()
		} else if len(tcpAddr.IP) == net.IPv6len {
			return "ipv6://" + tcpAddr.String()
		}
		return ""
	} else if unixAddr, ok := addr.(*net.UnixAddr); ok {
		return "unix://" + unixAddr.String()
	}
	return ""
}

func dialWithAddrAndTls(context context.Context, addr net.Addr, tls *tls.Config) (*grpc.ClientConn, error) {
	// Build dialing options.
	opts := make([]grpc.DialOption, 0, 3)
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.FailOnNonTempDialError(true))
	if tls == nil {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(tls)))
	}

	conn, err := grpc.DialContext(context, convertNetAddrToGRPCAddr(addr), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewGRPCClient(context context.Context, addr net.Addr, config *core.ClientConfig) (core.ClientInterface, error) {
	conn, err := dialWithAddrAndTls(context, addr, config.TLS)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		config:  config,
		client:  NewPluginInterfaceClient(conn),
		context: context,
		conn:    conn,
	}, nil
}

func (c *GRPCClient) Initialization() error {
	retErr, err := c.client.Initialization(c.context, &Empty{})
	if err != nil {
		return err
	}
	if retErr.Message != "" {
		return errors.New(retErr.Message)
	}
	return nil
}

func (c *GRPCClient) UnInitialization() error {
	retErr, err := c.client.UnInitialization(c.context, &Empty{})
	if err != nil {
		return err
	}
	if retErr.Message != "" {
		return errors.New(retErr.Message)
	}
	return nil
}

func (c *GRPCClient) Deploy(name string) (interface{}, error) {
	retErr, err := c.client.Deploy(c.context, &DeployRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	if retErr.Message != "" {
		return nil, errors.New(retErr.Message)
	}
	p, ok := c.config.Plugins[name]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type: %s", name)
	}
	return p.Client(c.context, c.conn)
}

func (c *GRPCClient) Ping(name string) error {
	client := grpc_health_v1.NewHealthClient(c.conn)
	_, err := client.Check(c.context, &grpc_health_v1.HealthCheckRequest{
		Service: name,
	})
	return err
}
