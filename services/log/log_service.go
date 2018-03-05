package log

import (
	"sync"
	"github.com/zhsyourai/URCF-engine/services"
)

type Service interface {
	services.ServiceLifeCycle
}

var instance *logService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &logService{
		}
	})
	return instance
}

type logService struct {
	services.InitHelper

}

func (s *logService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *logService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

