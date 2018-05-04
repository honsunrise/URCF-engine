package global_configuration

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/services"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"sync"
)

/*
`(...) yaml:"[<key>][,<flag1>[,<flag2>]]" (...)`

The following flags are currently supported:
omitempty    Only include the field if it's not set to the zero
             value for the type or to empty slices or maps.
             Zero valued structs will be omitted if all their public
             fields are zero, unless they implement an IsZero
             method (see the IsZeroer interface type), in which
             case the field will be included if that method returns true.

flow         Marshal using a flow style (useful for structs,
             sequences and maps).

inline       Inline the field, which must be a struct or a map,
             causing all of its fields or keys to be processed as if
             they were part of the outer struct. For maps, keys must
             not conflict with the yaml keys of other struct fields.
*/

type Rpc struct {
	Port uint16
}

type Sys struct {
	WorkPath     string `yaml:"work-path"`
	DatabasePath string `yaml:"database-path"`
	PluginPath   string `yaml:"plugin-path"`
	PluginWebs   string `yaml:"plugin-webs"`
}

type GlobalConfig struct {
	Rpc Rpc
	Sys Sys
}

type Service interface {
	services.ServiceLifeCycle
	Get() GlobalConfig
	Write(config *GlobalConfig) error
}

var instance *globalConfigService
var once sync.Once

func GetGlobalConfig() Service {
	once.Do(func() {
		instance = &globalConfigService{
			only: GlobalConfig{
				Rpc: Rpc{Port: 8228},
				Sys: Sys{WorkPath: "./", PluginPath: "./plugin", DatabasePath: "./database"},
			},
		}
	})
	return instance
}

type globalConfigService struct {
	services.InitHelper
	only       GlobalConfig
	configFile *os.File
	configBuf  []byte
	lock       sync.RWMutex
}

func (s *globalConfigService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		var err error
		s.configFile, err = os.Open(arguments[0].(string))
		if err != nil {
			log.Fatalf("error: %v", err)
			return err
		}
		s.configBuf, err = ioutil.ReadAll(s.configFile)
		if err != nil {
			log.Fatalf("error: %v", err)
			return err
		}
		err = yaml.Unmarshal(s.configBuf, &s.only)
		if err != nil {
			log.Fatalf("error: %v", err)
			return err
		}
		return nil
	})
}

func (s *globalConfigService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		err := s.configFile.Close()
		if err != nil {
			log.Fatalf("error: %v", err)
			return err
		}
		return nil
	})
}

func (s *globalConfigService) Get() GlobalConfig {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.only
}

func (s *globalConfigService) Write(config *GlobalConfig) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.only = *config
	s.configBuf, err = yaml.Marshal(&s.only)
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}
	_, err = s.configFile.Write(s.configBuf)
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}
	return
}
