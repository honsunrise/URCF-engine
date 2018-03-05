package netfilter

import (
	"sync"
	"github.com/zhsyourai/URCF-engine/services"
)

type Service interface {
	services.ServiceLifeCycle
}

var instance *netfilterService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &netfilterService{
		}
	})
	return instance
}

type netfilterService struct {
	services.InitHelper
}

func (s *netfilterService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *netfilterService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

