package account

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/zhsyourai/URCF-engine/repositories/account"
	"golang.org/x/crypto/bcrypt"
)

var testID = "__test" + fmt.Sprint(rand.Int())
var testPassword = "password" + fmt.Sprint(rand.Int())
var repo = account.NewAccountRepository()

func TestAccountService_Register(t *testing.T) {
	s := NewAccountService(repo)
	_, err := s.Register(testID, testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}

	a, err := s.GetByID(testID)
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	if a.ID != testID {
		t.Errorf("%s(%s)", "Register error", "Account id not equ")
	}

	err = bcrypt.CompareHashAndPassword(a.Password, []byte(testPassword))
	if err != nil {
		t.Errorf("%s(%s)", "Register error", "Password not match")
	}
}

func TestAccountService_GetAll(t *testing.T) {
	s := NewAccountService(repo)
	_, err := s.Register(testID+"1", testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	_, err = s.Register(testID+"2", testPassword, []string{"admin"})
	if err != nil {
		t.Errorf("%s(%s)", "Register error", fmt.Sprint(err))
	}
	_, err = s.Register(testID+"3", testPassword, []string{"admin"})
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
