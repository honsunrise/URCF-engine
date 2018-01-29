package repositories

import (
	"errors"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/google/uuid"
	"log"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"bytes"
	"encoding/gob"
	"reflect"
	"golang.org/x/crypto/argon2"
)

// AccountRepository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type AccountRepository interface {
	io.Closer
	insertAccount(models.Account) error
	findAccountByID(id string) (models.Account, error)
	deleteAccountByID(id string) (models.Account, error)
	updateAccountByID(id string, account map[string]interface{}) error
}

// NewAccountRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewAccountRepository() AccountRepository {
	db, err := leveldb.OpenFile("Account.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	return &accountBoltRepository{db}
}

// accountBoltRepository is a "AccountRepository"
// which manages the accounts using the memory data source (map).
type accountBoltRepository struct {
	db *leveldb.DB
}

func (r *accountBoltRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *accountBoltRepository) insertAccount(account models.Account) error {
	id := uuid.Must(uuid.NewRandom())
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(account)
	err := r.db.Put([]byte(id.String()), buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *accountBoltRepository) findAccountByID(id string) (account models.Account, err error) {
	value, err := r.db.Get([]byte(id),nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&account)
	if err != nil {
		return
	}
	return
}

func (r *accountBoltRepository) deleteAccountByID(id string) (account models.Account, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}
	value, err := trans.Get([]byte(id),nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&account)
	if err != nil {
		return
	}
	err = trans.Delete([]byte(id),nil)
	if err != nil {
		return
	}
	trans.Commit()
	return
}

func (r *accountBoltRepository) updateAccountByID(id string, account map[string]interface{}) error {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return err
	}

	value, err := trans.Get([]byte(id),nil)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	var originAccount models.Account
	err = dec.Decode(&originAccount)
	if err != nil {
		return err
	}
	s := reflect.ValueOf(originAccount).Elem()
	for k, v := range account {
		s.FieldByName(k).Set(reflect.ValueOf(v))
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(originAccount)
	trans.Put([]byte(id), buf.Bytes(), nil)
	trans.Commit()
	return nil
}