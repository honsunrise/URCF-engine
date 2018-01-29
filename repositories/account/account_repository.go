package account

import (
	"github.com/zhsyourai/URCF-engine/models"
	"log"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"bytes"
	"encoding/gob"
	"reflect"
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertAccount(models.Account) error
	FindAccountByID(id string) (models.Account, error)
	FindAll() ([]models.Account, error)
	DeleteAccountByID(id string) (models.Account, error)
	UpdateAccountByID(id string, account map[string]interface{}) error
}

// NewAccountRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewAccountRepository() Repository {
	db, err := leveldb.OpenFile("Account.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	return &accountRepository{db}
}

// accountRepository is a "Repository"
// which manages the accounts using the memory data source (map).
type accountRepository struct {
	db *leveldb.DB
}

func (r *accountRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *accountRepository) InsertAccount(account models.Account) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(account)
	if err != nil {
		return err
	}
	err = r.db.Put([]byte(account.ID), buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *accountRepository) FindAccountByID(id string) (account models.Account, err error) {
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

func (r *accountRepository) FindAll() (accounts []models.Account, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}

	iter := trans.NewIterator(nil, nil)
	for iter.Next() {
		var account models.Account
		dec := gob.NewDecoder(bytes.NewBuffer(iter.Value()))
		err = dec.Decode(&account)
		if err != nil {
			trans.Discard()
			return
		}
		accounts = append(accounts, account)
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		trans.Discard()
		return
	}
	err = trans.Commit()
	if err != nil {
		trans.Discard()
		return
	}
	return
}

func (r *accountRepository) DeleteAccountByID(id string) (account models.Account, err error) {
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
		trans.Discard()
		return
	}
	err = trans.Delete([]byte(id),nil)
	if err != nil {
		trans.Discard()
		return
	}
	err = trans.Commit()
	if err != nil {
		trans.Discard()
		return
	}
	return
}

func (r *accountRepository) UpdateAccountByID(id string, account map[string]interface{}) error {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return err
	}

	value, err := trans.Get([]byte(id),nil)
	if err != nil {
		trans.Discard()
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	var originAccount models.Account
	err = dec.Decode(&originAccount)
	if err != nil {
		trans.Discard()
		return err
	}
	s := reflect.ValueOf(&originAccount).Elem()
	for k, v := range account {
		s.FieldByName(k).Set(reflect.ValueOf(v))
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(originAccount)
	if err != nil {
		trans.Discard()
		return err
	}
	err = trans.Put([]byte(id), buf.Bytes(), nil)
	if err != nil {
		trans.Discard()
		return err
	}
	err = trans.Commit()
	if err != nil {
		trans.Discard()
		return err
	}
	return nil
}