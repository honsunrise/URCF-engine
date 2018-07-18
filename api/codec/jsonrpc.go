package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zhsyourai/URCF-engine/api"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	JSONRPCVersion         = "2.0"
	ServiceMethodSeparator = "."
)

type jsonrpcNotificationParam struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result,omitempty"`
}

type jsonrpcRequest struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Version string          `json:"jsonrpc"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *jsonrpcError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *jsonrpcError) ErrorCode() int {
	return err.Code
}

type jsonrpcErrorResponse struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
	Error  *jsonrpcError   `json:"error"`
}

type jsonrpcResponse struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
}

type jsonrpcRequests struct {
	reqs    []jsonrpcRequest
	isBatch bool
}

func (r *jsonrpcRequests) UnmarshalJSON(b []byte) error {
	var msg json.RawMessage
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}
	isArray := false
	for _, c := range msg {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		if c == '[' {
			isArray = true
		}
		break
	}
	if isArray {
		var result []jsonrpcRequest
		err := json.Unmarshal(b, &result)
		if err != nil {
			return err
		}
		r.reqs = result
		r.isBatch = true
	} else {
		var result jsonrpcRequest
		err := json.Unmarshal(b, &result)
		if err != nil {
			return err
		}
		r.reqs = []jsonrpcRequest{result}
		r.isBatch = false
	}
	return nil
}

type jsonrpcServerCodec struct {
	closer sync.Once
	closed chan interface{}
	decode *json.Decoder
	encode *json.Encoder
	rwc    io.ReadWriteCloser
}

func NewJsonRPCServerCodec(rwc io.ReadWriteCloser) api.ServerCodec {
	dec := json.NewDecoder(rwc)
	dec.UseNumber()
	return &jsonrpcServerCodec{
		closed: make(chan interface{}),
		encode: json.NewEncoder(rwc),
		decode: dec,
		rwc:    rwc,
	}
}

func (c *jsonrpcServerCodec) ReadRequest() ([]api.RPCRequest, bool, error) {
	var requests jsonrpcRequests
	err := c.decode.Decode(&requests)
	if err != nil {
		return nil, false, &api.InvalidMessageError{Message: err.Error()}
	}
	result := make([]api.RPCRequest, len(requests.reqs))
	for i, r := range requests.reqs {
		if err := checkReqId(r.ID); err != nil {
			return nil, false, &api.InvalidMessageError{Message: err.Error()}
		}

		id := &r.ID

		if len(r.Params) == 0 {
			result[i] = api.RPCRequest{ID: id, Params: nil}
		} else {
			dec := json.NewDecoder(bytes.NewReader(r.Params))
			if tok, _ := dec.Token(); tok != json.Delim('[') {
				return nil, false, &api.InvalidParamsError{Message: "non-array params"}
			}
			var params []json.RawMessage
			for i := 0; dec.More(); i++ {
				var v json.RawMessage
				if err := dec.Decode(&v); err != nil {
					return nil, false, &api.InvalidParamsError{
						Message: fmt.Sprintf("invalid argument %d: %v", i, err),
					}
				}
				params = append(params, v)
			}
			// Read end of args array.
			if _, err := dec.Token(); err != nil {
				return nil, false, &api.InvalidParamsError{Message: err.Error()}
			}
		}

		if elem := strings.Split(r.Method, ServiceMethodSeparator); len(elem) == 3 {
			result[i].Service, result[i].Executable, result[i].Method = elem[0], elem[1], elem[2]
		} else {
			result[i].Err = &api.MethodNotFoundError{Method: r.Method}
		}
	}
	return result, requests.isBatch, nil
}

func (c *jsonrpcServerCodec) ParsePosition(argTypes []reflect.Type, params []interface{}) ([]reflect.Value, error) {
	// Read args.
	args := make([]reflect.Value, 0, len(argTypes))
	for i := 0; i < len(argTypes); i++ {
		if i >= len(params) {
			return nil, &api.InvalidParamsError{
				Message: fmt.Sprintf("too many params request, only have %d", len(params)),
			}
		}

		argVal := reflect.New(argTypes[i])
		if err := json.Unmarshal(params[i].(json.RawMessage), argVal.Interface()); err != nil {
			return nil, &api.InvalidParamsError{Message: fmt.Sprintf("invalid argument %d: %v", i, err)}
		}
		if argVal.IsNil() && argTypes[i].Kind() != reflect.Ptr {
			return nil, &api.InvalidParamsError{Message: fmt.Sprintf("missing value for required argument %d", i)}
		}
		args = append(args, argVal.Elem())
	}

	// Set any missing args to nil.
	for i := len(params); i < len(argTypes); i++ {
		if argTypes[i].Kind() != reflect.Ptr {
			return nil, &api.InvalidParamsError{Message: fmt.Sprintf("missing value for required argument %d", i)}
		}
		args = append(args, reflect.Zero(argTypes[i]))
	}
	return args, nil
}

func (c *jsonrpcServerCodec) Write(responses []*api.RPCResponse, isBatch bool) error {
	return nil
}

func (c *jsonrpcServerCodec) Close() {
	c.closer.Do(func() {
		close(c.closed)
		c.rwc.Close()
	})
}

func (c *jsonrpcServerCodec) Closed() <-chan interface{} {
	return c.closed
}

func checkReqId(reqId json.RawMessage) error {
	if len(reqId) == 0 {
		return fmt.Errorf("missing request id")
	}
	if _, err := strconv.ParseFloat(string(reqId), 64); err == nil {
		return nil
	}
	var str string
	if err := json.Unmarshal(reqId, &str); err == nil {
		return nil
	}
	return fmt.Errorf("invalid request id")
}

type jsonrpcClientCodec struct {
	closer sync.Once
	closed chan interface{}
	decode *json.Decoder
	encode *json.Encoder
	rwc    io.ReadWriteCloser
}

func NewJsonRPCClientCodec(rwc io.ReadWriteCloser) api.ServerCodec {
	dec := json.NewDecoder(rwc)
	dec.UseNumber()
	return &jsonrpcServerCodec{
		closed: make(chan interface{}),
		encode: json.NewEncoder(rwc),
		decode: dec,
		rwc:    rwc,
	}
}

func (c *jsonrpcClientCodec) ReadResponse() ([]api.RPCResponse, bool, error) {
	panic("implement me")
}

func (c *jsonrpcClientCodec) ParsePosition(argTypes []reflect.Type, params []interface{}) ([]reflect.Value, error) {
	panic("implement me")
}

func (c *jsonrpcClientCodec) Write(request []*api.RPCRequest, isBatch bool) error {
	panic("implement me")
}

func (c *jsonrpcClientCodec) Close() {
	panic("implement me")
}

func (c *jsonrpcClientCodec) Closed() <-chan interface{} {
	panic("implement me")
}
