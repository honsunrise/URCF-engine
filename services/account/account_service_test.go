package account

import (
	"fmt"
	"math/rand"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var testUsername = "__test" + fmt.Sprint(rand.Int())
var testPassword = "password" + fmt.Sprint(rand.Int())

func TestAccountService_Register(t *testing.T) {
	s := GetInstance()
	_, err := s.Register(testUsername, testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}

	a, err := s.GetByUsername(testUsername)
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	if a.Username != testUsername {
		t.Errorf("%s(%s)", "Register error", "Account id not equ")
	}

	err = bcrypt.CompareHashAndPassword(a.Password, []byte(testPassword))
	if err != nil {
		t.Errorf("%s(%s)", "Register error", "Password not match")
	}
}

func TestAccountService_GetAll(t *testing.T) {
	s := GetInstance()
	_, err := s.Register(testUsername+"1", testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	_, err = s.Register(testUsername+"2", testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	_, err = s.Register(testUsername+"3", testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	accounts, err := s.GetAll()
	if err != nil {
		t.Errorf("%s(%s)", "GetAll error", fmt.Sprint(err))
	}
	if len(accounts) != 3 {
		t.Errorf("%s(%s)", "GetAll error", "Length 3 not equ ("+fmt.Sprint(len(accounts))+")")
	}
}
