package grpc

import "context"
import (
	"errors"
	grpc1 "google.golang.org/grpc"
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
