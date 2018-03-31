package log

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/services"
	"sync"
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
			debug: true,
		}
	})
	return instance
}

type logService struct {
	services.InitHelper
	debug bool
}

func (s *logService) GetLogger(name string) (*log.Entry, error) {
	logger := log.New()
	if s.debug {
		logger.SetLevel(log.DebugLevel)
	}
	return logger.WithField("name", name), nil
}

func (s *logService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		if s.debug {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	})
}

func (s *logService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}
