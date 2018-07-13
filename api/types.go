package api

import (
	"reflect"
)

type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

type call struct {
	rcvr          reflect.Value  // receiver of method
	method        reflect.Method // call
	argTypes      []reflect.Type // input argument types
	hasCtx        bool           // method's first argument is a context (not included in argTypes)
	hasError      bool           // err return idx, of -1 when method cannot return error
	isSubscribe   bool           // indication if the call is a subscription
	isUnsubscribe bool           // indication if the call is a subscription
}

type service struct {
	name          string        // name for service
	typ           reflect.Type  // receiver type
	callbacks     calls         // registered handlers
	subscriptions subscriptions // available subscriptions/notifications
}

type serviceRegistry map[string]*service // collection of services
type calls map[string]*call              // collection of RPC calls
type subscriptions map[string]*call      // collection of subscription calls

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
