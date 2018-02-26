package configuration

import "sync"

type Node struct {
	Value interface{}
}


func (*Node) Get(key string) Node {
	panic("implement me")
}

func (*Node) Put(key string, value interface{}) {
	panic("implement me")
}

func (*Node) Delete(key string) Node {
	panic("implement me")
}

type Service interface {
	Get(key string) Node
	Put(key string, value interface{})
	Delete(key string) Node
}

var rootNode *Node
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		rootNode = &Node{
		}
	})
	return rootNode
}
