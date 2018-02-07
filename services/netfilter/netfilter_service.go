package netfilter

import "sync"

type Service interface {
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
}
