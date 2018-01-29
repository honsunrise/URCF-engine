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

// PluginRepository handles the basic operations of a plugin entity/model.
// It's an interface in order to be testable, i.e a memory plugin repository or
// a connected to an sql database.
type PluginRepository interface {
	io.Closer
	InsertPlugin(plugin models.Plugin) error
	FindPluginByID(id string) (models.Plugin, error)
	DeletePluginByID(id string) (models.Plugin, error)
	UpdatePluginByID(id string, plugin map[string]interface{}) error
}

// NewPluginRepository returns a new plugin memory-based repository,
// the one and only repository type in our example.
func NewPluginRepository() PluginRepository {
	db, err := leveldb.OpenFile("Plugin.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	return &pluginRepository{db}
}

// pluginRepository is a "PluginRepository"
// which manages the plugins using the memory data source (map).
type pluginRepository struct {
	db *leveldb.DB
}

func (r *pluginRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *pluginRepository) InsertPlugin(plugin models.Plugin) error {
	id := uuid.Must(uuid.NewRandom())
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(plugin)
	err := r.db.Put([]byte(id.String()), buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *pluginRepository) FindPluginByID(id string) (plugin models.Plugin, err error) {
	value, err := r.db.Get([]byte(id),nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&plugin)
	if err != nil {
		return
	}
	return
}

func (r *pluginRepository) DeletePluginByID(id string) (plugin models.Plugin, err error) {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}
	value, err := trans.Get([]byte(id),nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&plugin)
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

func (r *pluginRepository) UpdatePluginByID(id string, plugin map[string]interface{}) error {
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return err
	}

	value, err := trans.Get([]byte(id),nil)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	var originPlugin models.Plugin
	err = dec.Decode(&originPlugin)
	if err != nil {
		return err
	}
	s := reflect.ValueOf(originPlugin).Elem()
	for k, v := range plugin {
		s.FieldByName(k).Set(reflect.ValueOf(v))
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(originPlugin)
	trans.Put([]byte(id), buf.Bytes(), nil)
	trans.Commit()
	return nil
}