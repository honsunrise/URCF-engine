package log

import (
	"sync"
	"github.com/zhsyourai/URCF-engine/services"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	services.ServiceLifeCycle
	GetLogger(name string) (*log.Entry, error)
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

func (s *logService) GetLogger(name string) (*log.Entry, error) {
	logger := log.New()
	logger.Out = w
	return logger.WithField("name", name), nil
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

