package api

import (
	"reflect"
)

type RPCRequest struct {
	Service    string
	Executable string
	Method     string
	ID         interface{}
	Params     interface{}
	Err        error
}

type RPCResponse struct {
	Service    string
	Executable string
	Method     string
	ID         interface{}
	Params     interface{}
	Err        error
	SubId      string
	Payload    interface{}
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
	// Read incoming response
	ReadResponse() ([]RPCResponse, bool, error)
	// parse incoming response
	ParsePosition(argTypes []reflect.Type, params []interface{}) ([]reflect.Value, error)
	// Write reply to client. Don't need set ID
	Write(request []*RPCRequest, isBatch bool) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
