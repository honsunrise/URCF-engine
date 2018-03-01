package configuration

import (
	"fmt"
	"math/rand"
	"testing"

)


func TestConfigurationService_Put(t *testing.T) {
	s := GetInstance()

	testKey := "test1." + fmt.Sprint(rand.Int())
	testValue := "value" + fmt.Sprint(rand.Int())
	err := s.Put(testKey, testValue)
	if err != nil {
		t.Errorf("%s(%s)", "Put error", fmt.Sprint(err))
		t.FailNow()
	}
	node, err := s.Get(testKey)
	if err != nil {
		t.Errorf("%s(%s)", "Get error", fmt.Sprint(err))
		t.FailNow()
	}
	if node.Key != testKey {
		t.Errorf("%s(%s)", "Put error", "Key not equ")
		t.FailNow()
	}

	testKey = "test2." + fmt.Sprint(rand.Int())
	testValue = "value" + fmt.Sprint(rand.Int())
	err = s.Put(testKey, testValue)
	if err != nil {
		t.Errorf("%s(%s)", "Put error", fmt.Sprint(err))
		t.FailNow()
	}

	node, err = s.Get(testKey)
	if err != nil {
		t.Errorf("%s(%s)", "Get error", fmt.Sprint(err))
		t.FailNow()
	}
	if node.Key != testKey {
		t.Errorf("%s(%s)", "Put error", "Key not equ")
		t.FailNow()
	}

	node, err = s.GetRoot()
	if err != nil {
		t.Errorf("%s(%s)", "Get all error", fmt.Sprint(err))
		t.FailNow()
	}
	nodes, err := node.GetAll()
	if err != nil {
		t.Errorf("%s(%s)", "Get all error", fmt.Sprint(err))
		t.FailNow()
	}
	if len(nodes) != 2 {
		t.Errorf("%s(%s)", "Get all error", "Length not equ")
		t.FailNow()
	}
}