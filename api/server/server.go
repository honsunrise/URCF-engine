package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/api"
	"gopkg.in/fatih/set.v0"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

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

type subscription struct {
	ID         string
	service    string
	executable string
	err        chan error // closed on unsubscribe
	exit       chan struct{}
}

type requestBound struct {
	request *api.RPCRequest
	params  []reflect.Value
	call    *call
	err     error
}

func (s *subscription) Err() <-chan error {
	return s.err
}

type Server struct {
	services serviceRegistry

	subMu  sync.RWMutex // guards subMap maps
	subMap map[string]*subscription

	run      int32
	codecsMu sync.Mutex
	codecs   *set.Set
}

func NewServer() *Server {
	server := &Server{
		services: make(serviceRegistry),
		codecs:   set.New(),
		run:      1,
	}

	rpcService := &RPCMetaService{server}
	server.RegisterName(MetadataApi, rpcService)

	return server
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	svc := new(service)
	svc.typ = reflect.TypeOf(rcvr)
	rcvrVal := reflect.ValueOf(rcvr)

	if name == "" {
		return fmt.Errorf("no service name for type %s", svc.typ.String())
	}
	if !isExported(reflect.Indirect(rcvrVal).Type().Name()) {
		return fmt.Errorf("%s is not exported", reflect.Indirect(rcvrVal).Type().Name())
	}

	methods, subscriptions := suitableMethods(rcvr)

	if regSvc, present := s.services[name]; present {
		if len(methods) == 0 && len(subscriptions) == 0 {
			return fmt.Errorf("service %q doesn't have any suitable methods/subscriptions to expose", name)
		}
		for _, m := range methods {
			regSvc.callbacks[formatName(m.method.Name)] = m
		}
		for _, s := range subscriptions {
			regSvc.subscriptions[formatName(s.method.Name)] = s
		}
		return nil
	}

	svc.name = name
	svc.callbacks, svc.subscriptions = methods, subscriptions

	if len(svc.callbacks) == 0 && len(svc.subscriptions) == 0 {
		return fmt.Errorf("service %q doesn't have any suitable methods/subscriptions to expose", name)
	}

	s.services[svc.name] = svc
	return nil
}

func (s *Server) ServeCodec(codec api.ServerCodec) error {
	defer codec.Close()
	return s.serveRequest(context.Background(), codec)
}

func (s *Server) serveRequest(ctx context.Context, codec api.ServerCodec) error {
	var pend sync.WaitGroup

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error(string(buf))
		}
		s.codecsMu.Lock()
		s.codecs.Remove(codec)
		s.codecsMu.Unlock()
	}()

	//	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.codecsMu.Lock()
	if atomic.LoadInt32(&s.run) != 1 { // server stopped
		s.codecsMu.Unlock()
		return &api.ShutdownError{}
	}
	s.codecs.Add(codec)
	s.codecsMu.Unlock()

	for atomic.LoadInt32(&s.run) == 1 {
		reqs, isBatch, err := s.readRequest(codec)
		if err != nil {
			log.Debug(fmt.Sprintf("read error %v\n", err))
			codec.Write([]*api.RPCResponse{{
				Err: &api.CallbackError{err.Error()},
			}}, false)
			pend.Wait()
			return nil
		}

		if atomic.LoadInt32(&s.run) != 1 {
			err = &api.ShutdownError{}
			resps := make([]*api.RPCResponse, len(reqs))
			for i, r := range reqs {
				resps[i] = &api.RPCResponse{
					ID:  &r.request.ID,
					Err: &api.CallbackError{Message: err.Error()},
				}
			}
			codec.Write(resps, isBatch)
			return nil
		}

		pend.Add(1)
		go func(reqs []*requestBound) {
			defer pend.Done()
			s.exec(ctx, codec, reqs, isBatch)
		}(reqs)
	}
	return nil
}

func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.Debug("RPC Server shutdown initialized")
		s.codecsMu.Lock()
		defer s.codecsMu.Unlock()
		s.codecs.Each(func(c interface{}) bool {
			c.(api.ServerCodec).Close()
			return true
		})
	}
}

func (s *Server) createSubscription(ctx context.Context, codec api.ServerCodec, req *requestBound) (*subscription, error) {
	args := []reflect.Value{req.call.rcvr, reflect.ValueOf(ctx)}
	args = append(args, req.params...)
	reply := req.call.method.Func.Call(args)

	if req.call.hasError && !reply[1].IsNil() {
		return nil, reply[1].Interface().(error)
	}

	sub := &subscription{ID: uuid.Must(uuid.NewRandom()).String(), err: make(chan error), exit: make(chan struct{})}
	s.subMu.Lock()
	s.subMap[sub.ID] = sub
	s.subMu.Unlock()

	go func() {
		for {
			cases := []reflect.SelectCase{
				{
					Dir:  reflect.SelectRecv,
					Chan: reply[0],
				},
				{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(sub.exit),
				},
			}
			switch index, value, recvOK := reflect.Select(cases); index {
			case 0:
				if recvOK == true {
					err := s.notify(codec, sub.ID, value)
					if err != nil {
						sub.err <- err
					}
				}
			case 1:
				break
			}
		}
	}()

	return sub, nil
}

func (s *Server) unsubscribe(id string) error {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	if sub, found := s.subMap[id]; found {
		close(sub.err)
		close(sub.exit)
		delete(s.subMap, id)
		return nil
	}
	return ErrSubscriptionNotFound
}

func (s *Server) notify(codec api.ServerCodec, id string, data interface{}) error {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	sub, ok := s.subMap[id]
	if ok {
		notification := &api.RPCResponse{
			SubId:      sub.ID,
			Service:    sub.service,
			Executable: sub.executable,
		}
		if err := codec.Write([]*api.RPCResponse{notification}, false); err != nil {
			codec.Close()
			return err
		}
	}
	return nil
}

func (s *Server) handle(ctx context.Context, codec api.ServerCodec, req *requestBound) *api.RPCResponse {
	if req.call.isUnsubscribe {
		if len(req.params) >= 1 && req.params[0].Kind() == reflect.String {
			subId := req.params[0].String()
			if err := s.unsubscribe(subId); err != nil {
				return &api.RPCResponse{ID: req.request.ID, Err: &api.CallbackError{Message: err.Error()}}
			} else {
				return &api.RPCResponse{ID: req.request.ID, Payload: true}
			}
		}
		return &api.RPCResponse{
			ID:  req.request.ID,
			Err: &api.InvalidParamsError{Message: "Expected subscription id as first argument"},
		}
	} else if req.call.isSubscribe {
		sub, err := s.createSubscription(ctx, codec, req)
		if err != nil {
			return &api.RPCResponse{ID: req.request.ID, Err: &api.CallbackError{Message: err.Error()}}
		}

		return &api.RPCResponse{ID: req.request.ID, Payload: sub.ID}
	} else {
		if len(req.params) != len(req.call.argTypes) {
			rpcErr := &api.InvalidParamsError{
				Message: fmt.Sprintf("%s[%s] expects %d parameters, but got %d",
					req.request.Service, req.request.Method, len(req.call.argTypes), len(req.params)),
			}
			return &api.RPCResponse{ID: req.request.ID, Err: rpcErr}
		}

		arguments := []reflect.Value{req.call.rcvr}
		if req.call.hasCtx {
			arguments = append(arguments, reflect.ValueOf(ctx))
		}
		if len(req.params) > 0 {
			arguments = append(arguments, req.params...)
		}

		// execute RPC method and return result
		reply := req.call.method.Func.Call(arguments)
		if len(reply) == 0 {
			return &api.RPCResponse{ID: req.request.ID, Payload: nil}
		}
		if req.call.hasError { // test if method returned an error
			if !reply[len(reply)-1].IsNil() {
				e := reply[len(reply)-1].Interface().(error)
				return &api.RPCResponse{ID: req.request.ID, Err: &api.CallbackError{Message: e.Error()}}
			}
		}
		return &api.RPCResponse{ID: req.request.ID, Payload: reply[0].Interface()}
	}
}

// exec executes the given requests and writes the result back using the codec.
// It will only write the response back when the last request is processed.
func (s *Server) exec(ctx context.Context, codec api.ServerCodec, requests []*requestBound, isBatch bool) {
	responses := make([]*api.RPCResponse, len(requests))
	for i, req := range requests {
		if req.err != nil {
			responses[i] = &api.RPCResponse{ID: req.request.ID, Err: req.err}
		} else {
			responses[i] = s.handle(ctx, codec, req)
		}
	}

	if err := codec.Write(responses, isBatch); err != nil {
		log.Error(fmt.Sprintf("%v\n", err))
		codec.Close()
	}
}

func (s *Server) readRequest(codec api.ServerCodec) ([]*requestBound, bool, error) {
	reqs, isBatch, err := codec.ReadRequest()
	if err != nil {
		return nil, false, err
	}

	requests := make([]*requestBound, len(reqs))

	// verify requests
	for i, r := range reqs {
		var ok bool
		var svc *service

		if r.Err != nil {
			requests[i] = &requestBound{request: &r, err: r.Err}
			continue
		}

		if svc, ok = s.services[r.Service]; !ok {
			requests[i] = &requestBound{request: &r, err: &api.MethodNotFoundError{
				Method: fmt.Sprintf("%s[%s]", r.Service, r.Method),
			}}
			continue
		}

		if call, ok := svc.subscriptions[r.Method]; ok {
			requests[i] = &requestBound{request: &r, call: call}
			if r.Params != nil && len(call.argTypes) > 0 {
				argTypes := []reflect.Type{reflect.TypeOf("")}
				argTypes = append(argTypes, call.argTypes...)
				if params, err := codec.ParsePosition(argTypes, r.Params.([]interface{})); err == nil {
					requests[i].params = params
				} else {
					requests[i].err = &api.InvalidParamsError{Message: err.Error()}
				}
			}
			continue
		}

		if call, ok := svc.callbacks[r.Method]; ok {
			requests[i] = &requestBound{request: &r, call: call}
			if r.Params != nil && len(call.argTypes) > 0 {
				if params, err := codec.ParsePosition(call.argTypes, r.Params.([]interface{})); err == nil {
					requests[i].params = params
				} else {
					requests[i].err = &api.InvalidParamsError{Message: err.Error()}
				}
			}
			continue
		}

		requests[i] = &requestBound{request: &r, err: &api.MethodNotFoundError{
			Method: fmt.Sprintf("%s[%s]", r.Service, r.Method),
		}}
	}

	return requests, isBatch, nil
}
