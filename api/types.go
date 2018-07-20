package api

import (
	"reflect"
)

type Decorator int

const (
	Subscribe Decorator = 1 << iota
	Unsubscribe
)

type RPCRequest struct {
	Service   string
	Method    string
	Decorator Decorator
	ID        interface{}
	Params    interface{}
	Err       error
}

type RPCResponse struct {
	Service string
	Method  string
	ID      interface{}
	Err     error
	SubId   string
	Payload interface{}
}

type ServerCodec interface {
	// Read incoming request
	ReadRequest() ([]RPCRequest, bool, error)
	// parse incoming request
	ParsePosition(argTypes []reflect.Type, params []interface{}) ([]reflect.Value, error)
	// Write reply to client.
	Write(responses []*RPCResponse, isBatch bool) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}

type ClientCodec interface {
	// get next id
	NextId() interface{}
	// Read incoming response
	ReadResponse() ([]RPCResponse, bool, error)
	// parse incoming response
	ParsePosition(argTypes []reflect.Type, params []interface{}) ([]reflect.Value, error)
	// Write reply to client.
	Write(request []*RPCRequest, isBatch bool) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
