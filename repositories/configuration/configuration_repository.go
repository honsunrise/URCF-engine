package configuration

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"path"
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertConfig(config models.Config) error
	FindConfigByKey(key string) (models.Config, error)
	FindAll() ([]models.Config, error)
	DeleteConfigByKey(key string) (models.Config, error)
	UpdateConfigByKey(key string, config map[string]interface{}) error
}

// NewConfigurationRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewConfigurationRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbFile := path.Join(confServ.Get().Sys.WorkPath, "database", "Configuration.db")
	db, err := leveldb.OpenFile(dbFile, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &configurationRepository{db}
}

// configurationRepository is a "Repository"
// which manages the accounts using the memory data source (map).
type configurationRepository struct {
	db *leveldb.DB
}

func (r *configurationRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *configurationRepository) InsertConfig(config models.Config) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(config)
	if err != nil {
		return err
	}
	err = r.db.Put([]byte(config.Key), buf.Bytes(), &opt.WriteOptions{
		NoWriteMerge: true,
		Sync: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *configurationRepository) FindConfigByKey(key string) (account models.Config, err error) {
	value, err := r.db.Get([]byte(key), nil)
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

func (r *configurationRepository) FindAll() (accounts []models.Config, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}

	iter := trans.NewIterator(nil, nil)
	for iter.Next() {
		var account models.Config
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

func (r *configurationRepository) DeleteConfigByKey(key string) (account models.Config, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}
	value, err := trans.Get([]byte(key), nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&account)
	if err != nil {
		trans.Discard()
		return
	}
	err = trans.Delete([]byte(key), nil)
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

func (r *configurationRepository) UpdateConfigByKey(key string, account map[string]interface{}) error {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return err
	}

	value, err := trans.Get([]byte(key), nil)
	if err != nil {
		trans.Discard()
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	var originConfig models.Config
	err = dec.Decode(&originConfig)
	if err != nil {
		trans.Discard()
		return err
	}
	s := reflect.ValueOf(&originConfig).Elem()
	for k, v := range account {
		s.FieldByName(k).Set(reflect.ValueOf(v))
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(originConfig)
	if err != nil {
		trans.Discard()
		return err
	}
	err = trans.Put([]byte(key), buf.Bytes(), nil)
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
