package async

import (
	"errors"
	"sync"
)

type ResultHandler interface {
	Handle(interface{})
}

type EmittableFunc func() interface{}

type (
	ResultFunc func(interface{})

	ErrFunc func(error)
)

var ErrResultCanNotBeError = errors.New("result can't be error")
var ErrResultMustBeError = errors.New("result must be error")
var ErrOneOfAKindHandler = errors.New("one of a kind handler")

func (handle ResultFunc) Handle(item interface{}) {
	switch item := item.(type) {
	case error:
		panic(ErrResultCanNotBeError)
	default:
		handle(item)
	}
}

func (handle ErrFunc) Handle(item interface{}) {
	switch item := item.(type) {
	case error:
		handle(item)
	default:
		panic(ErrResultMustBeError)
	}
}

type HandlerBundle struct {
	ResultHandler ResultFunc
	ErrHandler    ErrFunc
}

func (hb HandlerBundle) OnNext(item interface{}) {
	switch item := item.(type) {
	case error:
		return
	default:
		if hb.ResultHandler != nil {
			hb.ResultHandler(item)
		}
	}
}

func (hb HandlerBundle) OnError(err error) {
	if hb.ErrHandler != nil {
		hb.ErrHandler(err)
	}
}

func BundleResultHandler(handlers []ResultHandler) HandlerBundle {
	ob := HandlerBundle{}
	for _, handler := range handlers {
		switch handler := handler.(type) {
		case ResultFunc:
			if ob.ResultHandler != nil {
				panic(ErrOneOfAKindHandler)
			}
			ob.ResultHandler = handler
		case ErrFunc:
			if ob.ErrHandler != nil {
				panic(ErrOneOfAKindHandler)
			}
			ob.ErrHandler = handler
		}
	}

	return ob
}

type AsyncRet <-chan interface{}

// Subscribe subscribes an ResultHandler and returns a Subscription channel.
func (o AsyncRet) Subscribe(handlers ...ResultHandler) <-chan struct{} {
	done := make(chan struct{})

	ob := BundleResultHandler(handlers)

	go func() {
		for item := range o {
			switch item := item.(type) {
			case error:
				ob.OnError(item)
			default:
				ob.OnNext(item)
			}
		}

		done <- struct{}{}
		return
	}()

	return done
}

func Empty() AsyncRet {
	source := make(chan interface{})
	go func() {
		close(source)
	}()
	return AsyncRet(source)
}

func Just(item interface{}, items ...interface{}) AsyncRet {
	source := make(chan interface{})
	if len(items) > 0 {
		items = append([]interface{}{item}, items...)
	} else {
		items = []interface{}{item}
	}

	go func() {
		for _, item := range items {
			source <- item
		}
		close(source)
	}()

	return AsyncRet(source)
}

func From(f EmittableFunc, fs ...EmittableFunc) AsyncRet {
	if len(fs) > 0 {
		fs = append([]EmittableFunc{f}, fs...)
	} else {
		fs = []EmittableFunc{f}
	}

	source := make(chan interface{})

	var wg sync.WaitGroup
	for _, f := range fs {
		wg.Add(1)
		go func(f EmittableFunc) {
			source <- f()
			wg.Done()
		}(f)
	}

	// Wait in another goroutine to not block
	go func() {
		wg.Wait()
		close(source)
	}()

	return AsyncRet(source)
}
