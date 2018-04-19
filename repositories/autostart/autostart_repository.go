package autostart

import (
	"io"
	"log"
	"reflect"

	"database/sql"
	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"os"
	"path"
)

const (
	_CREATE_TABLE_SQL_ = `CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			priority INTEGER NOT NULL,
			start_delay INTEGER NOT NULL,
			stop_delay INTEGER NOT NULL,
			parallel BOOLEAN NOT NULL,
			enable BOOLEAN NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL
		)`

	_INSERT_SQL = `INSERT INTO autostarts(priority, start_delay, stop_delay, parallel, enable, name, cmd, args, workdir, env, option, create_time, update_time)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_EXIST_BY_ID_SQL = `SELECT EXISTS(SELECT id FROM autostarts WHERE id = ?)`

	_SELECT_ALL_SQL = `SELECT * FROM autostarts`

	_SELECT_BY_ID_SQL = `SELECT * FROM autostarts WHERE id = ?`

	_DELETE_BY_ID_SQL = `DELETE FROM autostarts WHERE id = ?`

	_DELETE_ALL_SQL = `DELETE FROM autostarts`

	_UPDATE_BY_ID_SQL = `UPDATE autostarts SET priority = ?, start_delay = ?, stop_delay = ?, parallel = ?, 
			enable = ?, name = ?, cmd = ?, args = ?, workdir = ?, env = ?, option = ?, 
			update_time = CURRENT_TIMESTAMP WHERE id = ?`
)

// Repository handles the basic operations of a AutoStart entity/model.
// It's an interface in order to be testable, i.e a memory AutoStart repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertAutoStart(autoStart *models.AutoStart) error
	FindAutoStartByID(id int64) (models.AutoStart, error)
	FindAll() ([]models.AutoStart, error)
	DeleteAutoStartByID(id int64) (models.AutoStart, error)
	DeleteAll() error
	UpdateAutoStartByID(id int64, fields map[string]interface{}) (models.AutoStart, error)
}

// NewAutostartRepository returns a new AutoStart memory-based repository,
// the one and only repository type in our example.
func NewAutostartRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbPath := path.Join(confServ.Get().Sys.WorkPath, "database")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0770)
	}
	dbFile := path.Join(dbPath, "Autostart.db")

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(_CREATE_TABLE_SQL_)
	if err != nil {
		log.Fatal(err)
	}
	return &autostartRepository{db}
}

// autostartRepository is a "Repository"
// which manages the AutoStarts using the memory data source (map).
type autostartRepository struct {
	db *sql.DB
}

func (r *autostartRepository) InsertAutoStart(autoStart *models.AutoStart) (err error) {
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

	result, err := tx.Exec(_INSERT_SQL, &autoStart.ID, &autoStart.Priority, &autoStart.StartDelay, &autoStart.StopDelay,
		&autoStart.Parallel, &autoStart.Enable, &autoStart.Name, &autoStart.Cmd, &autoStart.Args,
		&autoStart.WorkDir, &autoStart.Env, &autoStart.Option)
	if err != nil {
		return
	}
	autoStart.ID, err = result.LastInsertId()
	success = true
	return
}

func (r *autostartRepository) FindAutoStartByID(id int64) (autoStart models.AutoStart, err error) {
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

	err = tx.QueryRow(_SELECT_BY_ID_SQL, id).Scan(
		&autoStart.ID, &autoStart.Priority, &autoStart.StartDelay, &autoStart.StopDelay,
		&autoStart.Parallel, &autoStart.Enable, &autoStart.CreateTime, &autoStart.UpdateTime, &autoStart.Name,
		&autoStart.Cmd, &autoStart.Args, &autoStart.WorkDir, &autoStart.Env, &autoStart.Option)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *autostartRepository) FindAll() (autoStarts []models.AutoStart, err error) {
	autoStarts = make([]models.AutoStart, 0, 100)
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

	rows, err := tx.Query(_SELECT_ALL_SQL)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var autoStart models.AutoStart
		err = rows.Scan(
			&autoStart.ID, &autoStart.Priority, &autoStart.StartDelay, &autoStart.StopDelay,
			&autoStart.Parallel, &autoStart.Enable, &autoStart.CreateTime, &autoStart.UpdateTime, &autoStart.Name,
			&autoStart.Cmd, &autoStart.Args, &autoStart.WorkDir, &autoStart.Env, &autoStart.Option)
		if err != nil {
			return
		}
		autoStarts = append(autoStarts, autoStart)
	}
	success = true
	return
}

func (r *autostartRepository) DeleteAutoStartByID(id int64) (autoStart models.AutoStart, err error) {
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
	err = tx.QueryRow(_SELECT_BY_ID_SQL, id).Scan(
		&autoStart.ID, &autoStart.Priority, &autoStart.StartDelay, &autoStart.StopDelay,
		&autoStart.Parallel, &autoStart.Enable, &autoStart.CreateTime, &autoStart.UpdateTime, &autoStart.Name,
		&autoStart.Cmd, &autoStart.Args, &autoStart.WorkDir, &autoStart.Env, &autoStart.Option)
	if err != nil {
		return
	}

	_, err = tx.Exec(_DELETE_BY_ID_SQL, id)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *autostartRepository) UpdateAutoStartByID(id int64,
	fields map[string]interface{}) (autoStart models.AutoStart, err error) {
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
	err = tx.QueryRow(_SELECT_BY_ID_SQL, id).Scan(
		&autoStart.ID, &autoStart.Priority, &autoStart.StartDelay, &autoStart.StopDelay,
		&autoStart.Parallel, &autoStart.Enable, &autoStart.CreateTime, &autoStart.UpdateTime, &autoStart.Name,
		&autoStart.Cmd, &autoStart.Args, &autoStart.WorkDir, &autoStart.Env, &autoStart.Option)
	if err != nil {
		return
	}

	s := reflect.ValueOf(&autoStart).Elem()
	for k, v := range fields {
		field := s.FieldByName(k)
		if field.IsValid() {
			field.Set(reflect.ValueOf(v))
		} else {
			err = errors.New(fmt.Sprintf("field %s not exist", k))
			return
		}
	}

	_, err = tx.Exec(_UPDATE_BY_ID_SQL, &autoStart.Priority, &autoStart.StartDelay,
		&autoStart.StopDelay, &autoStart.Parallel, &autoStart.Enable, &autoStart.Name, &autoStart.Cmd, &autoStart.Args,
		&autoStart.WorkDir, &autoStart.Env, &autoStart.Option)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *autostartRepository) DeleteAll() (err error) {
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

func (r *autostartRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
