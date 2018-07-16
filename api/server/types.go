package server

import (
	"reflect"
)

type RPCRequest struct {
	service    string
	executable string
	method     string
	id         interface{}
	params     interface{}
	err        error
}

type RPCResponse struct {
	id      interface{}
	payload interface{}
}

type ErrorResponse struct {
	id  interface{}
	err error
}

type NotifyResponse struct {
	subId      string
	service    string
	executable string
	payload    interface{}
}

type ServerCodec interface {
	// Read and parse incoming request
	ReadRequest() ([]RPCRequest, bool, error)
	// Read and parse incoming request
	ParseArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, error)
	// Write reply to client.
	Write(responses []interface{}, isBatch bool) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
