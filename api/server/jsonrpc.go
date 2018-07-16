package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type jsonRequest struct {
	Method  string          `json:"method"`
	Version string          `json:"jsonrpc"`
	Id      json.RawMessage `json:"id,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRequests struct {
	reqs    []jsonRequest
	isBatch bool
}

func (r *jsonRequests) UnmarshalJSON(b []byte) error {
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
		var result []jsonRequest
		json.Unmarshal(b, &result)
		r.reqs = result
		r.isBatch = true
	} else {
		var result jsonRequest
		json.Unmarshal(b, &result)
		r.reqs = []jsonRequest{result}
		r.isBatch = false
	}
	return nil
}

type jsonSuccessResponse struct {
	Version string      `json:"jsonrpc"`
	Id      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type jsonErrResponse struct {
	Version string      `json:"jsonrpc"`
	Id      interface{} `json:"id,omitempty"`
	Error   jsonError   `json:"error"`
}

type jsonSubscription struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result,omitempty"`
}

type jsonNotification struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  jsonSubscription `json:"params"`
}

type jsonCodec struct {
	closer sync.Once
	closed chan interface{}
	decode *json.Decoder
	encode *json.Encoder
	rwc    io.ReadWriteCloser
}

func (err *jsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *jsonError) ErrorCode() int {
	return err.Code
}

func NewCodec(rwc io.ReadWriteCloser) ServerCodec {
	dec := json.NewDecoder(rwc)
	dec.UseNumber()
	return &jsonCodec{
		closed: make(chan interface{}),
		encode: json.NewEncoder(rwc),
		decode: dec,
		rwc:    rwc,
	}
}

func (c *jsonCodec) ReadRequest() ([]RPCRequest, bool, error) {
	var requests jsonRequests
	err := c.decode.Decode(&requests)
	if err != nil {
		return nil, false, &invalidMessageError{err.Error()}
	}
	result := make([]RPCRequest, len(requests.reqs))
	for i, r := range requests.reqs {
		if err := checkReqId(r.Id); err != nil {
			return nil, false, &invalidMessageError{err.Error()}
		}

		id := &r.Id

		if len(r.Params) == 0 {
			result[i] = RPCRequest{id: id, params: nil}
		} else {
			result[i] = RPCRequest{id: id, params: r.Params}
		}

		if elem := strings.Split(r.Method, ServiceMethodSeparator); len(elem) == 3 {
			result[i].service, result[i].executable, result[i].method = elem[0], elem[1], elem[2]
		} else {
			result[i].err = &methodNotFoundError{r.Method, ""}
		}
	}
	return result, requests.isBatch, nil
}

func (c *jsonCodec) ParseArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, error) {
	return parsePositionalArguments(params.(json.RawMessage), argTypes)
}

func (c *jsonCodec) Write(responses []interface{}, isBatch bool) error {
	return nil
}

func (c *jsonCodec) Close() {
	c.closer.Do(func() {
		close(c.closed)
		c.rwc.Close()
	})
}

func (c *jsonCodec) Closed() <-chan interface{} {
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

func parsePositionalArguments(rawArgs json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	// Read beginning of the args array.
	dec := json.NewDecoder(bytes.NewReader(rawArgs))
	if tok, _ := dec.Token(); tok != json.Delim('[') {
		return nil, &invalidParamsError{"non-array args"}
	}
	// Read args.
	args := make([]reflect.Value, 0, len(types))
	for i := 0; dec.More(); i++ {
		if i >= len(types) {
			return nil, &invalidParamsError{fmt.Sprintf("too many arguments, want at most %d", len(types))}
		}
		argVal := reflect.New(types[i])
		if err := dec.Decode(argVal.Interface()); err != nil {
			return nil, &invalidParamsError{fmt.Sprintf("invalid argument %d: %v", i, err)}
		}
		if argVal.IsNil() && types[i].Kind() != reflect.Ptr {
			return nil, &invalidParamsError{fmt.Sprintf("missing value for required argument %d", i)}
		}
		args = append(args, argVal.Elem())
	}
	// Read end of args array.
	if _, err := dec.Token(); err != nil {
		return nil, &invalidParamsError{err.Error()}
	}
	// Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		if types[i].Kind() != reflect.Ptr {
			return nil, &invalidParamsError{fmt.Sprintf("missing value for required argument %d", i)}
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}
