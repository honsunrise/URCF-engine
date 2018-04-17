package core

import (
	"context"
)

type ClientInterface interface {
	Initialization() error

	UnInitialization() error

	Deploy(name string) (interface{}, error)

	Ping(name string) error
}

type ClientInstanceInterface interface {
	Instance(ctx context.Context, conn interface{}) (interface{}, error)
}
