package configuration

import (
	"io"
	"log"
	"reflect"

	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"os"
	"path"
)

const (
	_CREATE_TABLE_SQL_ = `CREATE TABLE IF NOT EXISTS configs (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL,
            expires INTEGER NOT NULL
		)`

	_INSERT_SQL = `INSERT INTO configs(key, value, expires, create_time, update_time)
			VALUES(?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_SELECT_BY_KEY_SQL = `SELECT * FROM configs WHERE key = ?`

	_SELECT_ALL_SQL = `SELECT * FROM configs`

	_COUNT_ALL_SQL = `SELECT COUNT(*) as count FROM configs`

	_DELETE_BY_KEY_SQL = `DELETE FROM configs WHERE key = ?`

	_DELETE_ALL_SQL = `DELETE FROM configs`

	_UPDATE_BY_KEY_SQL = `UPDATE configs SET value = ?, expires = ?, update_time = CURRENT_TIMESTAMP WHERE key = ?`
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertConfig(config *models.Config) error
	FindConfigByKey(key string) (models.Config, error)
	FindAll(page uint32, size uint32, sorts []repositories.Sort) ([]models.Config, error)
	CountAll() (int64, error)
	DeleteConfigByKey(key string) (models.Config, error)
	DeleteAll() error
	UpdateConfigByKey(key string, fields map[string]interface{}) (config models.Config, err error)
}

// NewConfigurationRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewConfigurationRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbPath := path.Join(confServ.Get().Sys.WorkPath, "database")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0770)
	}
	dbFile := path.Join(dbPath, "Configuration.db")

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(_CREATE_TABLE_SQL_)
	if err != nil {
		log.Fatal(err)
	}
	return &configurationRepository{OrderPaging: &repositories.OrderPaging{
		MaxSize: 100,
		CanOrderFields: map[string]repositories.Order{
			"key":         repositories.ASC | repositories.DESC,
			"expires":     repositories.ASC | repositories.DESC,
			"update_time": repositories.ASC | repositories.DESC,
			"create_time": repositories.ASC | repositories.DESC,
		},
	}, db: db}
}

// configurationRepository is a "Repository"
// which manages the accounts using the memory data source (map).
type configurationRepository struct {
	*repositories.OrderPaging
	db *sql.DB
}

func (r *configurationRepository) InsertConfig(config *models.Config) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.Exec(_INSERT_SQL, &config.Key, &config.Value, &config.Expires)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *configurationRepository) FindConfigByKey(key string) (config models.Config, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = tx.QueryRow(_SELECT_BY_KEY_SQL, key).Scan(
		&config.Key, &config.Value, &config.CreateTime, &config.UpdateTime, &config.Expires)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *configurationRepository) FindAll(page uint32, size uint32,
	sorts []repositories.Sort) (configs []models.Config, err error) {
	configs = make([]models.Config, 0, 100)
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()

	paSoStr, err := r.BuildPagingOrder(page, size, sorts)
	if err != nil {
		return
	}
	rows, err := tx.Query(_SELECT_ALL_SQL + paSoStr)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var config models.Config
		err = rows.Scan(&config.Key, &config.Value, &config.CreateTime, &config.UpdateTime, &config.Expires)
		if err != nil {
			return
		}
		configs = append(configs, config)
	}
	success = true
	return
}

func (r *configurationRepository) CountAll() (count int64, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = tx.QueryRow(_COUNT_ALL_SQL).Scan(&count)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *configurationRepository) DeleteConfigByKey(key string) (config models.Config, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = tx.QueryRow(_SELECT_BY_KEY_SQL, key).Scan(
		&config.Key, &config.Value, &config.CreateTime, &config.UpdateTime, &config.Expires)
	if err != nil {
		return
	}

	_, err = tx.Exec(_DELETE_BY_KEY_SQL, key)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *configurationRepository) DeleteAll() (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.Exec(_DELETE_ALL_SQL)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *configurationRepository) UpdateConfigByKey(key string,
	fields map[string]interface{}) (config models.Config, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = tx.QueryRow(_SELECT_BY_KEY_SQL, key).Scan(
		&config.Key, &config.Value, &config.CreateTime, &config.UpdateTime, &config.Expires)
	if err != nil {
		return
	}

	s := reflect.ValueOf(&config).Elem()
	for k, v := range fields {
		field := s.FieldByName(k)
		if field.IsValid() {
			field.Set(reflect.ValueOf(v))
		} else {
			err = errors.New(fmt.Sprintf("field %s not exist", k))
			return
		}
	}

	_, err = tx.Exec(_UPDATE_BY_KEY_SQL, config.Value, config.Expires)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *configurationRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
