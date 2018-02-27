package configuration

import (
	"sync"
	"github.com/zhsyourai/URCF-engine/repositories/configuration"
	"github.com/zhsyourai/URCF-engine/models"
	"time"
	"strings"
)

var startOf2018 = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

type Node struct {
	models.Config
	Child   []*Node
	Parent  *Node
	service Service
}

func (n *Node) Get(key string) (*Node, error) {
	for _, e := range n.Child {
		if e.Key == key {
			return e, nil
		}
	}
	return n.service.Get(key)
}

func (n *Node) Put(key string, value interface{}) error {
	panic("implement me")
}

func (n *Node) Delete(key string) *Node {
	panic("implement me")
}

type Service interface {
	Get(key string) (*Node, error)
	Put(key string, value interface{}) error
	Delete(key string) (*Node, error)
}

type configurationService struct {
	repo     configuration.Repository
	rootNode *Node
}

func (s *configurationService) Get(key string) (*Node, error) {
	allPath := strings.Split(key, ".")
	parentPath := ""
	parentNode := s.rootNode
	var currentNode *Node
	for _, path := range allPath {
		currentPath := parentPath + path
		exist := false
		for _, e := range parentNode.Child {
			if e.Key == currentPath {
				currentNode = e
				exist = true
				break
			}
		}
		if !exist {
			currentConfig, err := s.repo.FindConfigByKey(currentPath)
			if err != nil {
				if err.Error() != "leveldb: not found" {
					return &Node{}, err
				} else {
					currentConfig = models.Config{
						Key:        currentPath,
						Value:      nil,
						CreateDate: startOf2018,
						UpdateDate: startOf2018,
						Scope:      "",
						Expires:    time.Duration(-1),
					}
				}
			}
			currentNode = &Node{
				Parent:  parentNode,
				Child:   make([]*Node, 0, 10),
				service: s,
				Config:  currentConfig,
			}
			parentNode.Child = append(parentNode.Child, currentNode)
		}
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
	for _, path := range allPath {
		currentPath := parentPath + path
		exist := false
		for _, e := range parentNode.Child {
			if e.Key == currentPath {
				currentNode = e
				exist = true
				break
			}
		}
		if !exist {
			currentConfig, err := s.repo.FindConfigByKey(currentPath)
			if err != nil {
				if err.Error() != "leveldb: not found" {
					return err
				} else {
					currentConfig = models.Config{
						Key:        currentPath,
						Value:      nil,
						CreateDate: startOf2018,
						UpdateDate: startOf2018,
						Scope:      "",
						Expires:    time.Duration(-1),
					}
					if key == currentPath {
						currentConfig = models.Config{
							Key:        key,
							Value:      value,
							CreateDate: now,
							UpdateDate: now,
							Expires:    time.Duration(-1),
							Scope:      "",
						}
						err := s.repo.InsertConfig(currentConfig)
						if err != nil {
							return err
						}
						return nil
					}
				}
			}
			currentNode = &Node{
				Parent:  parentNode,
				Child:   make([]*Node, 0, 10),
				service: s,
				Config:  currentConfig,
			}
			parentNode.Child = append(parentNode.Child, currentNode)

		}
		currentNode.Value = value
		err := s.repo.InsertConfig(currentNode.Config)
		if err != nil {
			return err
		}
		parentPath = currentPath + "."
		parentNode = currentNode
	}
	return nil
}

func (s *configurationService) Delete(key string) (*Node, error) {
	config, err := s.repo.DeleteConfigByKey(key)
	if err != nil {
		return &Node{}, err
	}
	return &Node{
		service: s,
		Config:  config,
	}, nil
}

var service *configurationService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		service = &configurationService{
			repo: configuration.NewConfigurationRepository(),
			rootNode: &Node{
				Parent:  nil,
				Child:   make([]*Node, 0, 100),
				service: service,
				Config: models.Config{
					Key:        ".",
					Value:      nil,
					CreateDate: startOf2018,
					UpdateDate: startOf2018,
					Scope:      "",
					Expires:    time.Duration(-1),
				},
			},
		}
	})
	return service
}
