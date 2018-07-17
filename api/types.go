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
	Params     interface{}
	Err        error
	ID         interface{}
	Payload    interface{}
}

type ErrorResponse struct {
	ID  interface{}
	Err error
}

type NotifyResponse struct {
	SubId      string
	Service    string
	Executable string
	Payload    interface{}
}

type ServerCodec interface {
	// Read incoming request
	ReadRequest() ([]RPCRequest, bool, error)
	// parse incoming request
	ParsePosition(argTypes []reflect.Type, params []interface{}) ([]reflect.Value, error)
	// Write reply to client.
	Write(responses []interface{}, isBatch bool) error
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
	// Write reply to client.
	Write(request []interface{}, isBatch bool) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
