package configuration

import "sync"

type Service interface {
}

var instance *configurationService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &configurationService{
		}
	})
	return instance
}

type configurationService struct {
}
