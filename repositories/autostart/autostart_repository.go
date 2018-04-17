package autostart

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"path"
)

// Repository handles the basic operations of a AutoStart entity/model.
// It's an interface in order to be testable, i.e a memory AutoStart repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertAutoStart(autoStart *models.AutoStart) error
	FindAutoStartByID(id string) (models.AutoStart, error)
	FindAll() ([]models.AutoStart, error)
	DeleteAutoStartByID(id string) (models.AutoStart, error)
	UpdateAutoStartByID(id string, AutoStart map[string]interface{}) error
}

// NewAutostartRepository returns a new AutoStart memory-based repository,
// the one and only repository type in our example.
func NewAutostartRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbFile := path.Join(confServ.Get().Sys.WorkPath, "database", "AutoStart.db")
	db, err := leveldb.OpenFile(dbFile, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &autostartRepository{db}
}

// autostartRepository is a "Repository"
// which manages the AutoStarts using the memory data source (map).
type autostartRepository struct {
	db *leveldb.DB
}

func (r *autostartRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *autostartRepository) InsertAutoStart(autoStart *models.AutoStart) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(autoStart)
	if err != nil {
		return err
	}
	err = r.db.Put([]byte(autoStart.ID), buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *autostartRepository) FindAutoStartByID(id string) (autoStart models.AutoStart, err error) {
	value, err := r.db.Get([]byte(id), nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&autoStart)
	if err != nil {
		return
	}
	return
}

func (r *autostartRepository) FindAll() (autoStarts []models.AutoStart, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}

	autoStarts = make([]models.AutoStart, 0, 10)
	iter := trans.NewIterator(nil, nil)
	for iter.Next() {
		var autoStart models.AutoStart
		dec := gob.NewDecoder(bytes.NewBuffer(iter.Value()))
		err = dec.Decode(&autoStart)
		if err != nil {
			trans.Discard()
			return
		}
		autoStarts = append(autoStarts, autoStart)
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

func (r *autostartRepository) DeleteAutoStartByID(id string) (autoStart models.AutoStart, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}
	value, err := trans.Get([]byte(id), nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&autoStart)
	if err != nil {
		trans.Discard()
		return
	}
	err = trans.Delete([]byte(id), nil)
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

func (r *autostartRepository) UpdateAutoStartByID(id string, autoStart map[string]interface{}) error {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return err
	}

	value, err := trans.Get([]byte(id), nil)
	if err != nil {
		trans.Discard()
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	var originAutoStart models.AutoStart
	err = dec.Decode(&originAutoStart)
	if err != nil {
		trans.Discard()
		return err
	}
	s := reflect.ValueOf(&originAutoStart).Elem()
	for k, v := range autoStart {
		s.FieldByName(k).Set(reflect.ValueOf(v))
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(originAutoStart)
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
