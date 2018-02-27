package configuration

import (
	"fmt"
	"math/rand"
	"testing"

)


func TestConfigurationService_Put(t *testing.T) {
	s := GetInstance()

	testKey := "test." + fmt.Sprint(rand.Int())
	testValue := "value" + fmt.Sprint(rand.Int())
	err := s.Put(testKey, testValue)
	if err != nil {
		t.Errorf("%s(%s)", "Put error", fmt.Sprint(err))
	}
	node, err := s.Get(testKey)
	if err != nil {
		t.Errorf("%s(%s)", "Put error", fmt.Sprint(err))
	}
	if node.Key != testKey {
		t.Errorf("%s(%s)", "Put error", "Key not equ")
	}

	testValue = "value" + fmt.Sprint(rand.Int())
	err = s.Put(testKey, testValue)
	if err != nil {
		t.Errorf("%s(%s)", "Put error", fmt.Sprint(err))
	}

	node, err = s.Get(testKey)
	if err != nil {
		t.Errorf("%s(%s)", "Put error", fmt.Sprint(err))
	}
	if node.Key != testKey {
		t.Errorf("%s(%s)", "Put error", "Key not equ")
	}
}