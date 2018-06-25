package netfilter

import (
	"github.com/prometheus/common/log"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services"
	"sync"
)

type Service interface {
	services.ServiceLifeCycle
	ListAll(table string, chain string) ([]models.IptableRule, error)
	ListChains(table string) ([]string, error)
}

var instance *netfilterService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		iptables, err := New()
		if err != nil {
			panic(err)
		}
		instance = &netfilterService{
			iptables: iptables,
		}
	})
	return instance
}

type netfilterService struct {
	services.InitHelper
	iptables *IPTables
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

func (s *netfilterService) ListAll(table string, chain string) ([]models.IptableRule, error) {
	result, err := s.iptables.List(table, chain)
	log.Info(result)
	return nil, err
}

func (s *netfilterService) ListChains(table string) ([]string, error) {
	return s.iptables.ListChains(table)
}
