package services

import (
	"github.com/zhsyourai/teddy-backend/uaa/models"
	"github.com/zhsyourai/teddy-backend/uaa/repositories"
)

// AccountService handles some of the CRUID operations of the account datamodel.
// It depends on a account repository for its actions.
// It's here to decouple the data source from the higher level compoments.
// As a result a different repository type can be used with the same logic without any aditional changes.
// It's an interface and it's used as interface everywhere
// because we may need to change or try an experimental different domain logic at the future.
type AccountService interface {
	GetAll() []models.Account
	GetByID(id int64) (models.Account, bool)
	DeleteByID(id int64) bool
	UpdatePosterAndGenreByID(id int64, poster string, genre string) (models.Account, error)
}

// NewAccountService returns the default account service.
func NewAccountService(repo repositories.AccountRepository) AccountService {
	return &accountService{
		repo: repo,
	}
}

type accountService struct {
	repo repositories.AccountRepository
}

// GetAll returns all accounts.
func (s *accountService) GetAll() []models.Account {
	return s.repo.SelectMany(func(_ models.Account) bool {
		return true
	}, -1)
}

// GetByID returns a account based on its id.
func (s *accountService) GetByID(id int64) (models.Account, bool) {
	return s.repo.Select(func(m models.Account) bool {
		return m.ID == id
	})
}

// UpdatePosterAndGenreByID updates a account's poster and genre.
func (s *accountService) UpdatePosterAndGenreByID(id int64, poster string, genre string) (models.Account, error) {
	// update the account and return it.
	return s.repo.InsertOrUpdate(models.Account{
		ID:     id,
		Poster: poster,
		Genre:  genre,
	})
}

// DeleteByID deletes a account by its id.
//
// Returns true if deleted otherwise false.
func (s *accountService) DeleteByID(id int64) bool {
	return s.repo.Delete(func(m models.Account) bool {
		return m.ID == id
	}, 1)
}
