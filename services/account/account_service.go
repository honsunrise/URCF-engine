package account

import (
	"sync"
	"time"

	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/account"
	"golang.org/x/crypto/bcrypt"
	"github.com/zhsyourai/URCF-engine/services"
	"github.com/kataras/iris/core/errors"
	log "github.com/sirupsen/logrus"
)

var UserNotFoundErr = errors.New("user not found")
var OldPasswordNotCorrectErr = errors.New("old password not correct")
var PasswordModifyErr = errors.New("password modify error")

type Service interface {
	services.ServiceLifeCycle
	GetAll() ([]models.Account, error)
	GetByID(id string) (models.Account, error)
	DeleteByID(id string) (models.Account, error)
	Register(id string, password string, role []string) (models.Account, error)
	Verify(id string, password string) (models.Account, error)
	ChangePassword(id string, oldPassword string, newPassword string) error
}

var instance *accountService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &accountService{
			repo: account.NewAccountRepository(),
		}
	})
	return instance
}

type accountService struct {
	services.InitHelper
	repo account.Repository
}

func (s *accountService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *accountService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

func (s *accountService) Register(username string, password string, role []string) (account models.Account, err error) {
	now := time.Now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	account = models.Account{
		ID:         username,
		Password:   hashedPassword,
		Role:       role,
		CreateDate: now,
		UpdateDate: now,
		Enabled:    true,
	}
	err = s.repo.InsertAccount(account)
	if err != nil {
		return
	}
	return
}

func (s *accountService) GetAll() ([]models.Account, error) {
	accs, err := s.repo.FindAll()
	if err != nil {
		log.Error(err)
		return []models.Account{}, err
	}
	return accs, nil
}

func (s *accountService) GetByID(id string) (models.Account, error) {
	acc, err := s.repo.FindAccountByID(id)
	if err != nil {
		log.Error(err)
		return models.Account{}, UserNotFoundErr
	}
	return acc, nil
}

func (s *accountService) DeleteByID(id string) (models.Account, error) {
	acc, err := s.repo.DeleteAccountByID(id)
	if err != nil {
		log.Error(err)
		return models.Account{}, UserNotFoundErr
	}
	return acc, nil
}

func (s *accountService) Verify(id string, password string) (models.Account, error) {
	acc, err := s.repo.FindAccountByID(id)
	if err != nil {
		log.Error(err)
		return models.Account{}, UserNotFoundErr
	}
	err = bcrypt.CompareHashAndPassword(acc.Password, []byte(password))
	if err != nil {
		log.Error(err)
		return models.Account{}, UserNotFoundErr
	}
	return acc, nil
}

func (s *accountService) ChangePassword(id string, oldPassword string, newPassword string) error {
	acc, err := s.repo.FindAccountByID(id)
	if err != nil {
		log.Error(err)
		return UserNotFoundErr
	}
	err = bcrypt.CompareHashAndPassword(acc.Password, []byte(oldPassword))
	if err != nil {
		log.Error(err)
		return OldPasswordNotCorrectErr
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		return PasswordModifyErr
	}
	err = s.repo.UpdateAccountByID(id, map[string]interface{} {
		"Password": hashedPassword,
	})
	if err != nil {
		log.Error(err)
		return PasswordModifyErr
	}
	return nil
}
