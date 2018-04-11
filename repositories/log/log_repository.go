package log

import (
	"bytes"
	"encoding/gob"
	"io"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"path"
	"encoding/binary"
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
	DeleteLogByID(id uint64) (models.Log, error)
	DeleteAll() error
}

var _INDEX_NAME = []byte("~~index~~")

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
	indexBytes, err := trans.Get(_INDEX_NAME, nil)
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

	err = trans.Put(idBytes, buf.Bytes(), &opt.WriteOptions{
		NoWriteMerge: true,
		Sync:         true,
	})
	if err != nil {
		return 0, err
	}

	err = trans.Put(_INDEX_NAME, idBytes, &opt.WriteOptions{
		NoWriteMerge: true,
		Sync:         true,
	})
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
	value, err := trans.Get(idBytes, nil)
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

	iter := trans.NewIterator(nil, nil)
	for iter.Next() {
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
	value, err := trans.Get(idBytes, nil)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	err = dec.Decode(&log)
	if err != nil {
		trans.Discard()
		return
	}
	err = trans.Delete(idBytes, nil)
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

	var needDelete [][]byte
	iter := trans.NewIterator(nil, nil)
	for iter.Next() {
		needDelete = append(needDelete, iter.Key())
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		trans.Discard()
		return
	}

	for _, e := range needDelete {
		err = trans.Delete(e, nil)
		if err != nil {
			trans.Discard()
			return
		}
	}

	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, 0)

	err = trans.Put(_INDEX_NAME, idBytes, &opt.WriteOptions{
		NoWriteMerge: true,
		Sync:         true,
	})
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

func (r *logRepository) Close() error {
	r.l.Lock()
	defer r.l.Unlock()
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
