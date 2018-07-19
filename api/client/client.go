package client

import (
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/api"
	"reflect"
	"sync"
	"time"
)

var (
	ErrClientQuit                = errors.New("client is closed")
	ErrNoResult                  = errors.New("no Result in JSON-RPC response")
	ErrSubscriptionQueueOverflow = errors.New("subscription queue overflow")
	ErrNotificationsUnsupported  = errors.New("notifications unsupported")
)

const (
	subscribeTimeout = 5 * time.Second
	callTimeout      = 5 * time.Second
)

const (
	// Subscriptions are removed when the subscriber cannot keep up.
	//
	// This can be worked around by supplying a channel with sufficiently sized buffer,
	// but this can be inconvenient and hard to explain in the docs. Another issue with
	// buffered channels is that the buffer is static even though it might not be needed
	// most of the time.
	//
	// The approach taken here is to maintain a per-subscription linked list buffer
	// shrinks on demand. If the buffer reaches the size below, the subscription is
	// dropped.
	maxClientSubscriptionBuffer = 20000
)

type Client struct {
	codec     api.ClientCodec
	close     chan struct{}
	closeDone chan struct{}                 // closed when client quits
	rbChan    chan *requestBound            // for registering response IDs
	respWait  map[interface{}]*requestBound // active rbChan
	subs      map[interface{}]*subscription // active subscriptions
}

type requestBound struct {
	isBatch  bool
	requests []*api.RPCRequest
	err      error
	resp     chan *api.RPCResponse
	sub      *subscription
}

func (op *requestBound) wait(ctx context.Context) (*api.RPCResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-op.resp:
		return resp, op.err
	}
}

func NewClientWithCodec(codec api.ClientCodec) *Client {
	client := &Client{
		codec: codec,
	}
	go client.dispatch()
	return client
}

func (c *Client) SupportedModules() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	result, err := c.CallContext(ctx, "Meta", "List")
	return result.(map[string]string), err
}

func (c *Client) Close() {
	select {
	case c.close <- struct{}{}:
		<-c.closeDone
	case <-c.closeDone:
	}
}

func (c *Client) Call(result interface{}, service string, method string, params ...interface{}) (interface{}, error) {
	ctx := context.Background()
	return c.CallContext(ctx, result, service, method, params...)
}

func (c *Client) CallContext(ctx context.Context, result interface{}, service string, method string, params ...interface{}) (interface{}, error) {
	request := &api.RPCRequest{
		Service: service,
		Method:  method,
		Params:  params,
	}
	rb := &requestBound{requests: []*api.RPCRequest{request}, resp: make(chan *api.RPCResponse, 1)}

	err := c.send(ctx, rb)
	if err != nil {
		return nil, err
	}

	// dispatch has accepted the request and will close the channel when it quits.
	switch resp, err := rb.wait(ctx); {
	case err != nil:
		return nil, err
	case resp.Error != nil:
		return resp.Error
	case len(resp.Result) == 0:
		return ErrNoResult
	default:
		return json.Unmarshal(resp.Result, &result)
	}
}

type BatchItem struct {
	Service string
	Method  string
	Params  []interface{}
	Result  interface{}
}

type BatchBuilder struct {
	client *Client
	ctx    context.Context
	items  []BatchItem
}

func (b *BatchBuilder) Call(result interface{}, service string, method string, args ...interface{}) *BatchBuilder {
	return nil
}

func (b *BatchBuilder) Submit() error {
	rb := &requestBound{
		requests: make([]*api.RPCRequest, len(b.items)),
		resp:     make(chan *api.RPCResponse, len(b.items)),
	}
	for i, elem := range b.items {
		request := &api.RPCRequest{
			Service: elem.Service,
			Method:  elem.Method,
			Params:  elem.Params,
		}
		rb.requests[i] = request
	}

	err := b.client.send(b.ctx, rb)

	// Wait for all responses to come back.
	for n := 0; n < len(b.items) && err == nil; n++ {
		var resp *api.RPCResponse
		resp, err = rb.wait(b.ctx)
		if err != nil {
			break
		}
		// Find the element corresponding to this response.
		// The element is guaranteed to be present because dispatch
		// only sends valid IDs to our channel.
		var elem *BatchItem
		for i := range rb.requests {
			if bytes.Equal(rb.requests[i].ID, resp.ID) {
				elem = &b.items[i]
				break
			}
		}
		if resp.Err != nil {
			elem.Error = resp.Err
			continue
		}
		b.client.codec.ParsePosition([]reflect.Type{}, []interface{}{resp.Payload})
		if len() == 0 {
			elem.Error = ErrNoResult
			continue
		}
		elem.Error = json.Unmarshal(resp.Result, elem.Result)
	}
	return err
}

func (c *Client) Batch() *BatchBuilder {
	ctx := context.Background()
	return c.BatchContext(ctx)
}

func (c *Client) BatchContext(ctx context.Context) *BatchBuilder {
	return &BatchBuilder{
		client: c,
		ctx:    ctx,
		items:  make([]BatchItem, 10),
	}
}

func (c *Client) Subscribe(ctx context.Context, channel chan interface{},
	service string, method string, params ...interface{}) (Subscription, error) {
	request := &api.RPCRequest{
		Service: service,
		Method:  method,
		Params:  params,
	}
	rb := &requestBound{
		requests: []*api.RPCRequest{request},
		resp:     make(chan *api.RPCResponse),
		sub:      newClientSubscription(c, method, channel),
	}

	if err := c.send(ctx, rb); err != nil {
		return nil, err
	}
	if _, err := rb.wait(ctx); err != nil {
		return nil, err
	}
	return rb.sub, nil
}

func (c *Client) send(ctx context.Context, rb *requestBound) error {
	select {
	case c.rbChan <- rb:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closeDone:
		return ErrClientQuit
	}
}

func (c *Client) dispatch() {
	var (
		reading  = true
		readErr  chan error
		readResp chan *api.RPCResponse
	)
	defer close(c.closeDone)

	go func() {
		for {
			rs, _, err := c.codec.ReadResponse()
			if err != nil {
				readErr <- err
				return
			}
			for _, msg := range rs {
				readResp <- &msg
			}
		}
	}()

	defer func() {
		c.closeRequestOps(ErrClientQuit)
		if reading {
			// Empty read channels until read is dead.
			for {
				select {
				case <-readResp:
				case <-readErr:
					return
				}
			}
		}
	}()

	for {
		select {
		case <-c.close:
			return

		case resp := <-readResp:
			c.handleResponse(resp)

		case err := <-readErr:
			c.closeRequestOps(err)
			reading = false

		case rb := <-c.rbChan:
			err := c.codec.Write(rb.requests, rb.isBatch)
			if err != nil {
				c.closeRequestOps(err)
				reading = false
			} else {
				for _, req := range rb.requests {
					c.respWait[req.ID] = rb
				}
			}
		}
	}
}

// closeRequestOps unblocks pending send ops and active subscriptions.
func (c *Client) closeRequestOps(err error) {
	didClose := make(map[*requestBound]bool)

	for id, op := range c.respWait {
		// Remove the op so that later calls will not close op.resp again.
		delete(c.respWait, id)

		if !didClose[op] {
			op.err = err
			close(op.resp)
			didClose[op] = true
		}
	}
	for id, sub := range c.subs {
		delete(c.subs, id)
		sub.quitWithError(err, false)
	}
}

func (c *Client) handleResponse(resp *api.RPCResponse) {
	if resp.SubId != "" {
		if c.subs[resp.SubId] != nil {
			c.subs[resp.SubId].deliver(resp.Payload)
		}
	} else {
		rb := c.respWait[resp.ID]
		if rb == nil {
			log.Debug("unsolicited response", "resp", resp)
			return
		}
		delete(c.respWait, resp.ID)

		if resp.Err != nil {
			rb.err = resp.Err
			return
		}

		if rb.sub == nil {
			rb.resp <- resp
			return
		}

		defer close(rb.resp)

		if value, err := c.codec.ParsePosition([]reflect.Type{reflect.TypeOf(rb.sub.subId)},
			[]interface{}{resp.Payload}); err == nil {
			rb.sub.subId = value[0].String()
			go rb.sub.start()
			c.subs[rb.sub.subId] = rb.sub
		} else {
			rb.err = err
		}
	}
}

// Subscriptions.

type Subscription interface {
	Err() <-chan error
	Chan() <-chan interface{}
	Unsubscribe()
}

type subscription struct {
	client   *Client
	channel  chan interface{}
	chanType reflect.Type
	method   string
	subId    string
	in       chan interface{}

	quitOnce sync.Once     // ensures quit is closed once
	quit     chan struct{} // quit is closed when the subscription exits
	errOnce  sync.Once     // ensures err is closed once
	err      chan error
}

func newClientSubscription(c *Client, method string, channel chan interface{}) *subscription {
	chanType := reflect.TypeOf(channel)
	sub := &subscription{
		client:   c,
		method:   method,
		channel:  channel,
		chanType: chanType.Elem(),
		quit:     make(chan struct{}),
		err:      make(chan error, 1),
		in:       make(chan interface{}),
	}
	return sub
}

func (sub *subscription) Err() <-chan error {
	return sub.err
}

func (sub *subscription) Unsubscribe() {
	sub.quitWithError(nil, true)
	sub.errOnce.Do(func() { close(sub.err) })
}

func (sub *subscription) Chan() <-chan interface{} {
	return sub.channel
}

func (sub *subscription) quitWithError(err error, unsubscribeServer bool) {
	sub.quitOnce.Do(func() {
		// The dispatch loop won't be able to execute the unsubscribe call
		// if it is blocked on deliver. Close sub.quit first because it
		// unblocks deliver.
		close(sub.quit)
		if unsubscribeServer {
			var result interface{}
			sub.client.Call(&result, sub.method+unsubscribeMethodSuffix, sub.subId)
		}
		if err != nil {
			if err == ErrClientQuit {
				err = nil // Adhere to subscription semantics.
			}
			sub.err <- err
		}
	})
}

func (sub *subscription) deliver(result interface{}) (ok bool) {
	select {
	case sub.in <- result:
		return true
	case <-sub.quit:
		return false
	}
}

func (sub *subscription) start() {
	if err := sub.forward(); err != nil {
		sub.quitWithError(err, true)
	} else {
		sub.quitWithError(nil, false)
	}

}

func (sub *subscription) forward() error {
	cases := []reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(sub.quit)},
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(sub.in)},
		{Dir: reflect.SelectSend, Chan: reflect.ValueOf(sub.channel)},
	}
	buffer := list.New()
	defer buffer.Init()
	for {
		var chosen int
		var recv reflect.Value
		if buffer.Len() == 0 {
			// Idle, omit send case.
			chosen, recv, _ = reflect.Select(cases[:2])
		} else {
			// Non-empty buffer, send the first queued item.
			cases[2].Send = reflect.ValueOf(buffer.Front().Value)
			chosen, recv, _ = reflect.Select(cases)
		}

		switch chosen {
		case 0: // <-sub.quit
			return nil
		case 1: // <-sub.in
			val := reflect.New(sub.chanType)
			err := json.Unmarshal(recv.Interface().(json.RawMessage), val.Interface())
			if err != nil {
				return err
			}
			if buffer.Len() == maxClientSubscriptionBuffer {
				return ErrSubscriptionQueueOverflow
			}
			buffer.PushBack(val.Elem().Interface())
		case 2: // sub.channel<-
			cases[2].Send = reflect.Value{}
			buffer.Remove(buffer.Front())
		}
	}
}
