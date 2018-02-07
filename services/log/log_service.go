package log

import "sync"

type Service interface {
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
}
