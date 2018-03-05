package account

import (
	"sync"
	"time"

	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/account"
	"golang.org/x/crypto/bcrypt"
	"github.com/zhsyourai/URCF-engine/services"
)

type Service interface {
	services.ServiceLifeCycle
	GetAll() ([]models.Account, error)
	GetByID(id string) (models.Account, error)
	DeleteByID(id string) (models.Account, error)
	Register(id string, password string, role []string) (models.Account, error)
	Verify(id string, password string) (models.Account, error)
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
	return s.repo.FindAll()
}

func (s *accountService) GetByID(id string) (models.Account, error) {
	return s.repo.FindAccountByID(id)
}

func (s *accountService) DeleteByID(id string) (models.Account, error) {
	return s.repo.DeleteAccountByID(id)
}

func (s *accountService) Verify(id string, password string) (account models.Account, err error) {
	return
}
