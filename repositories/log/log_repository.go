package log

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"io"
	"path"
	"sync"
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertLog(log models.Log) (uint64, error)
	FindLogByID(id uint64) (models.Log, error)
	FindAll() ([]models.Log, error)
	Count() (uint64, error)
	DeleteLogByID(id uint64) (models.Log, error)
	DeleteAll() error
}

var _INDEX_NAME = []byte("~~index~~")

var _WRITE_OPTIONS = &opt.WriteOptions{
	NoWriteMerge: true,
	Sync:         true,
}

var _READ_OPTIONS = &opt.ReadOptions{
	DontFillCache: true,
	Strict:        opt.DefaultStrict,
}

func isReservedKey(key []byte) bool {
	if bytes.Compare(key, _INDEX_NAME) == 0 {
		return true
	}
	return false
}

// NewLogRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewLogRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbFile := path.Join(confServ.Get().Sys.WorkPath, "database", "Log.db")
	db, err := leveldb.OpenFile(dbFile, nil)
	if err != nil {
		panic(err)
	}
	trans, err := db.OpenTransaction()
	if err != nil {
		panic(err)
	}
	var index = uint64(0)
	indexBytes, err := trans.Get(_INDEX_NAME, _READ_OPTIONS)
	if err != nil {
		if err.Error() != "leveldb: not found" {
			panic(err)
		}
	} else {
		index = binary.LittleEndian.Uint64(indexBytes)
	}
	err = trans.Commit()
	if err != nil {
		trans.Discard()
		panic(err)
	}
	return &logRepository{db: db, index: index}
}

// logRepository is a "Repository"
// which manages the accounts using the memory data source (map).
type logRepository struct {
	db    *leveldb.DB
	index uint64
	l     sync.RWMutex
}

func (r *logRepository) Count() (uint64, error) {
	//TODO: Not to implement
	return 100, nil
}

func (r *logRepository) InsertLog(log models.Log) (uint64, error) {
	r.l.Lock()
	defer r.l.Unlock()
	trans, err := r.db.OpenTransaction()
	if err != nil {
		trans.Discard()
		return 0, err
	}
	success := false
	lastIndex := r.index

	defer func() {
		if !success {
			r.index = lastIndex
			trans.Discard()
			return
		}
	}()

	if log.ID == 0 {
		r.index++
		log.ID = r.index
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(log)
	if err != nil {
		return 0, err
	}

	idBytes := make([]byte, 8)

	binary.LittleEndian.PutUint64(idBytes, log.ID)

	err = trans.Put(idBytes, buf.Bytes(), _WRITE_OPTIONS)
	if err != nil {
		return 0, err
	}

	err = trans.Put(_INDEX_NAME, idBytes, _WRITE_OPTIONS)
	if err != nil {
		return 0, err
	}

	err = trans.Commit()
	if err != nil {
		return 0, err
	}
	success = true
	return log.ID, nil
}

func (r *logRepository) FindLogByID(id uint64) (log models.Log, err error) {
	r.l.RLock()
	defer r.l.RUnlock()
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}

	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, id)
	value, err := trans.Get(idBytes, _READ_OPTIONS)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&log)
	if err != nil {
		return
	}

	err = trans.Commit()
	if err != nil {
		trans.Discard()
		return
	}
	return
}

func (r *logRepository) FindAll() (logs []models.Log, err error) {
	r.l.RLock()
	defer r.l.RUnlock()
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}

	logs = make([]models.Log, 0, 100)
	iter := trans.NewIterator(nil, _READ_OPTIONS)
	for iter.Next() {
		if isReservedKey(iter.Key()) {
			continue
		}
		var log models.Log
		dec := gob.NewDecoder(bytes.NewBuffer(iter.Value()))
		err = dec.Decode(&log)
		if err != nil {
			trans.Discard()
			return
		}
		logs = append(logs, log)
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

func (r *logRepository) DeleteLogByID(id uint64) (log models.Log, err error) {
	r.l.Lock()
	defer r.l.Unlock()
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}
	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, id)
	value, err := trans.Get(idBytes, _READ_OPTIONS)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&log)
	if err != nil {
		trans.Discard()
		return
	}
	err = trans.Delete(idBytes, _WRITE_OPTIONS)
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

func (r *logRepository) DeleteAll() (err error) {
	r.l.Lock()
	defer r.l.Unlock()
	trans, err := r.db.OpenTransaction()
	if err != nil {
		return
	}

	iter := trans.NewIterator(nil, _READ_OPTIONS)
	for iter.Next() {
		err = trans.Delete(iter.Key(), _WRITE_OPTIONS)
		if err != nil {
			iter.Release()
			trans.Discard()
			return
		}
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		trans.Discard()
		return
	}

	r.index = 0

	err = trans.Commit()
	if err != nil {
		trans.Discard()
		return
	}
	return
}

func (r *logRepository) Close() error {
	r.l.Lock()
	defer r.l.Unlock()
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
