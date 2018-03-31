package services

import (
	"errors"
	"sync"
	"sync/atomic"
)

type ServiceLifeCycle interface {
	Initialize(arguments ...interface{}) error
	UnInitialize(arguments ...interface{}) error
}

type InitHelper struct {
	init uint32
	lock sync.RWMutex
}

func (s *InitHelper) CallInitialize(f func() error) error {
	if atomic.LoadUint32(&s.init) == 1 {
		return errors.New("already initialize")
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.init == 0 {
		defer atomic.StoreUint32(&s.init, 1)
		return f()
	}
	return errors.New("already initialize")
}

func (s *InitHelper) CallUnInitialize(f func() error) error {
	if atomic.LoadUint32(&s.init) == 0 {
		return errors.New("already uninitialize")
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.init == 1 {
		defer atomic.StoreUint32(&s.init, 0)
		return f()
	}
	return errors.New("already uninitialize")
}
