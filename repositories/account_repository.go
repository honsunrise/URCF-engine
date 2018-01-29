package repositories

import (
	"errors"
	"sync"
	"github.com/zhsyourai/teddy-backend/uaa/models"
)

// Query represents the visitor and action queries.
type Query func(models.Account) bool

// AccountRepository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type AccountRepository interface {
	Exec(query Query, action Query, limit int, mode int) (ok bool)

	Select(query Query) (account models.Account, found bool)
	SelectMany(query Query, limit int) (results []models.Account)

	InsertOrUpdate(account models.Account) (updatedAccount models.Account, err error)
	Delete(query Query, limit int) (deleted bool)
}

// NewAccountRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewAccountRepository(source map[int64]models.Account) AccountRepository {
	return &accountMemoryRepository{source: source}
}

// accountMemoryRepository is a "AccountRepository"
// which manages the accounts using the memory data source (map).
type accountMemoryRepository struct {
	source map[int64]models.Account
	mu     sync.RWMutex
}

const (
	// ReadOnlyMode will RLock(read) the data .
	ReadOnlyMode = iota
	// ReadWriteMode will Lock(read/write) the data.
	ReadWriteMode 
)

func (r *accountMemoryRepository) Exec(query Query, action Query, actionLimit int, mode int) (ok bool) {
	loops := 0

	if mode == ReadOnlyMode {
		r.mu.RLock()
		defer r.mu.RUnlock()
	} else {
		r.mu.Lock()
		defer r.mu.Unlock()
	}

	for _, account := range r.source {
		ok = query(account)
		if ok {
			if action(account) {
				loops++
				if actionLimit >= loops {
					break // break
				}
			}
		}
	}

	return
}

// Select receives a query function
// which is fired for every single account model inside
// our imaginary data source.
// When that function returns true then it stops the iteration.
//
// It returns the query's return last known "found" value
// and the last known account model
// to help callers to reduce the LOC.
//
// It's actually a simple but very clever prototype function
// I'm using everywhere since I firstly think of it,
// hope you'll find it very useful as well.
func (r *accountMemoryRepository) Select(query Query) (account models.Account, found bool) {
	found = r.Exec(query, func(m models.Account) bool {
		account = m
		return true
	}, 1, ReadOnlyMode)

	// set an empty models.Account if not found at all.
	if !found {
		account = models.Account{}
	}

	return
}

// SelectMany same as Select but returns one or more models.Account as a slice.
// If limit <=0 then it returns everything.
func (r *accountMemoryRepository) SelectMany(query Query, limit int) (results []models.Account) {
	r.Exec(query, func(m models.Account) bool {
		results = append(results, m)
		return true
	}, limit, ReadOnlyMode)

	return
}

// InsertOrUpdate adds or updates a account to the (memory) storage.
//
// Returns the new account and an error if any.
func (r *accountMemoryRepository) InsertOrUpdate(account models.Account) (models.Account, error) {
	id := account.ID

	if id == 0 { // Create new action
		var lastID int64
		// find the biggest ID in order to not have duplications
		// in productions apps you can use a third-party
		// library to generate a UUID as string.
		r.mu.RLock()
		for _, item := range r.source {
			if item.ID > lastID {
				lastID = item.ID
			}
		}
		r.mu.RUnlock()

		id = lastID + 1
		account.ID = id

		// map-specific thing
		r.mu.Lock()
		r.source[id] = account
		r.mu.Unlock()

		return account, nil
	}

	// Update action based on the account.ID,
	// here we will allow updating the poster and genre if not empty.
	// Alternatively we could do pure replace instead:
	// r.source[id] = account
	// and comment the code below;
	current, exists := r.Select(func(m models.Account) bool {
		return m.ID == id
	})

	if !exists { // ID is not a real one, return an error.
		return models.Account{}, errors.New("failed to update a nonexistent account")
	}

	// or comment these and r.source[id] = m for pure replace
	if account.Poster != "" {
		current.Poster = account.Poster
	}

	if account.Genre != "" {
		current.Genre = account.Genre
	}

	// map-specific thing
	r.mu.Lock()
	r.source[id] = current
	r.mu.Unlock()

	return account, nil
}

func (r *accountMemoryRepository) Delete(query Query, limit int) bool {
	return r.Exec(query, func(m models.Account) bool {
		delete(r.source, m.ID)
		return true
	}, limit, ReadWriteMode)
}
