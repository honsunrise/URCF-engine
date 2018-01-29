package account

import (
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories/account"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetAll() ([]models.Account, error)
	GetByID(id string) (models.Account, error)
	DeleteByID(id string) error
	Register(id string, password string, role []string) (models.Account, error)
}

func NewAccountService(repo account.Repository) Service {
	return &accountService{repo:repo}
}

type accountService struct {
	repo account.Repository
}

func (s *accountService) Register(username string, password string, role []string) (account models.Account, err error) {
	now := time.Now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	account = models.Account{
		ID: username,
		Password: hashedPassword,
		Role: role,
		CreateDate: now,
		UpdateDate: now,
		Enabled: true,
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

func (s *accountService) DeleteByID(id string) (err error) {
	_, err = s.repo.DeleteAccountByID(id)
	return
}
