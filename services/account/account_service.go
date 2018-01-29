package account

import (
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
)

type Service interface {
	GetAll() ([]models.Account, error)
	GetByID(id string) (models.Account, error)
	DeleteByID(id string) error
	Register(id string, password string, role []string) (models.Account, error)
}

func NewAccountService(repo repositories.AccountRepository) Service {
	return &accountService{repo:repo}
}

type accountService struct {
	repo repositories.AccountRepository
}

func (s *accountService) Register(username string, password string, role []string) (models.Account, error) {
	return models.Account{}, nil
}

func (s *accountService) GetAll() ([]models.Account, error) {
	return nil, nil
}

func (s *accountService) GetByID(id string) (models.Account, error) {
	return models.Account{}, nil
}

func (s *accountService) DeleteByID(id string) error {
	return nil
}
