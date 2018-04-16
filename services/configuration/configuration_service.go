package configuration

import (
	"errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/configuration"
	"github.com/zhsyourai/URCF-engine/services"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var startOf2018 = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

type Node struct {
	models.Config
	child   sync.Map
	parent  *Node
	service Service
}

func (n *Node) Get(key string) (*Node, error) {
	return n.service.Get(n.parent.Key + "." + key)
}

func (n *Node) GetAll() (nodes []*Node, err error) {
	n.child.Range(func(key, value interface{}) bool {
		nodes = append(nodes, value.(*Node))
		return true
	})
	return
}

func (n *Node) Put(key string, value interface{}) error {
	return n.service.Put(n.parent.Key+"."+key, value)
}

func (n *Node) Delete(key string) (*Node, error) {
	return n.service.Delete(n.parent.Key + "." + key)
}

type Service interface {
	services.ServiceLifeCycle
	Get(key string) (*Node, error)
	GetRoot() (*Node, error)
	Put(key string, value interface{}) error
	Delete(key string) (*Node, error)
}

type configurationService struct {
	services.InitHelper
	repo     configuration.Repository
	rootNode *Node
	syncFlag atomic.Value
}

func (s *configurationService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *configurationService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

func (s *configurationService) sync() error {
	configs, err := s.repo.FindAll()
	if err != nil {
		return err
	}
	for _, conf := range configs {
		allPath := strings.Split(conf.Key, ".")
		parentPath := ""
		parentNode := s.rootNode
		var currentNode *Node
		for i, path := range allPath {
			currentPath := parentPath + path
			tmp, exist := parentNode.child.Load(currentPath)
			if exist {
				currentNode = tmp.(*Node)
				if i == len(allPath)-1 {
					currentNode.Config = conf
				}
			} else {
				var currentConfig models.Config
				if i == len(allPath)-1 {
					currentConfig = conf
				} else {
					currentConfig = models.Config{
						Key:        currentPath,
						Value:      nil,
						CreateDate: startOf2018,
						UpdateDate: startOf2018,
						Expires:    time.Duration(-1),
					}
				}
				currentNode = &Node{
					parent:  parentNode,
					service: s,
					Config:  currentConfig,
				}
				parentNode.child.Store(currentPath, currentNode)
			}
			parentPath = currentPath + "."
			parentNode = currentNode
		}
	}
	s.syncFlag.Store(true)
	return nil
}

func (s *configurationService) GetRoot() (*Node, error) {
	return s.rootNode, nil
}

func (s *configurationService) Get(key string) (*Node, error) {
	allPath := strings.Split(key, ".")
	parentPath := ""
	parentNode := s.rootNode
	var currentNode *Node
	for _, path := range allPath {
		currentPath := parentPath + path
		tmp, exist := parentNode.child.Load(currentPath)
		if !exist {
			return nil, errors.New("configuration: key not exist")
		}
		currentNode = tmp.(*Node)
		parentPath = currentPath + "."
		parentNode = currentNode
	}
	return currentNode, nil
}

func (s *configurationService) Put(key string, value interface{}) error {
	allPath := strings.Split(key, ".")
	parentPath := ""
	parentNode := s.rootNode
	now := time.Now()
	var currentNode *Node
	for i, path := range allPath {
		currentPath := parentPath + path
		tmp, exist := parentNode.child.Load(currentPath)
		if exist {
			currentNode = tmp.(*Node)
			if i == len(allPath)-1 {
				currentNode.Value = value
				err := s.repo.InsertConfig(currentNode.Config)
				if err != nil {
					return err
				}
			}
		} else {
			var currentConfig models.Config
			if i == len(allPath)-1 {
				currentConfig = models.Config{
					Key:        key,
					Value:      value,
					CreateDate: now,
					UpdateDate: now,
					Expires:    time.Duration(-1),
				}
				err := s.repo.InsertConfig(currentConfig)
				if err != nil {
					return err
				}

			} else {
				currentConfig = models.Config{
					Key:        currentPath,
					Value:      nil,
					CreateDate: startOf2018,
					UpdateDate: startOf2018,
					Expires:    time.Duration(-1),
				}
			}
			currentNode = &Node{
				parent:  parentNode,
				service: s,
				Config:  currentConfig,
			}
			parentNode.child.Store(currentPath, currentNode)
		}
		parentPath = currentPath + "."
		parentNode = currentNode
	}
	return nil
}

func (s *configurationService) Delete(key string) (*Node, error) {
	allPath := strings.Split(key, ".")
	parentPath := ""
	parentNode := s.rootNode
	var currentNode *Node
	var exist = false
	for _, path := range allPath {
		currentPath := parentPath + path
		var tmp interface{}
		tmp, exist = parentNode.child.Load(currentPath)
		if !exist {
			return nil, errors.New("configuration: key not exist")
		}
		currentNode = tmp.(*Node)
		parentPath = currentPath + "."
		parentNode = currentNode
	}
	currentNode.parent.child.Delete(key)
	_, err := s.repo.DeleteConfigByKey(key)
	if err != nil {
		return &Node{}, err
	}
	return currentNode, nil
}

var service *configurationService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		service = &configurationService{
			repo: configuration.NewConfigurationRepository(),
			rootNode: &Node{
				parent:  nil,
				service: service,
				Config: models.Config{
					Key:        "_urcf_root_",
					Value:      nil,
					CreateDate: startOf2018,
					UpdateDate: startOf2018,
					Expires:    time.Duration(-1),
				},
			},
		}
		service.syncFlag.Store(false)
		service.sync()
	})
	return service
}
