package grpc

import "context"
import (
	grpc1 "google.golang.org/grpc"
	"errors"
)

type CommandPlugin struct {
}

func (cp *CommandPlugin) Client(ctx context.Context, conn interface{}) (interface{}, error) {
	realConn, ok := conn.(*grpc1.ClientConn)
	if !ok {
		return nil, errors.New("conn must be grpc.ClientConn")
	}
	return NewCommandInterfaceClient(realConn), nil
}
