package log

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"io"
	"log"
	"os"
	"path"
)

const (
	_CREATE_TABLE_SQL_ = `CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			message TEXT NOT NULL,
			level TEXT NOT NULL,
			create_time DATETIME NOT NULL
		)`

	_INSERT_SQL = `INSERT INTO logs(name, message, level, create_time)
			VALUES(?, ?, ?, CURRENT_TIMESTAMP)`

	_SELECT_BY_ID_SQL = `SELECT * FROM logs WHERE id = ?`

	_SELECT_BY_NAME_SQL = `SELECT * FROM logs WHERE name = ?`

	_SELECT_ALL_SQL = `SELECT * FROM logs`

	_COUNT_ALL_SQL = `SELECT COUNT(*) as count FROM logs`

	_COUNT_BY_NAME_SQL = `SELECT COUNT(*) as count FROM logs WHERE name = ?`

	_DELETE_BY_NAME_SQL = `DELETE FROM logs WHERE name = ?`

	_DELETE_BY_ID_SQL = `DELETE FROM logs WHERE id = ?`

	_DELETE_ALL_SQL = `DELETE FROM logs`
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertLog(log *models.Log) error
	FindLogByID(id int64) (models.Log, error)
	FindLogByName(name string, page uint32, size uint32, sorts []repositories.Sort) ([]models.Log, error)
	FindAll(page uint32, size uint32, sorts []repositories.Sort) ([]models.Log, error)
	CountAll() (int64, error)
	CountByName(name string) (int64, error)
	DeleteLogByID(id int64) (models.Log, error)
	DeleteLogByName(name string) error
	DeleteAll() error
}

// NewLogRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewLogRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbPath := confServ.Get().Sys.DatabasePath
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0770)
	}
	dbFile := path.Join(dbPath, "Log.db")

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(_CREATE_TABLE_SQL_)
	if err != nil {
		log.Fatal(err)
	}
	return &logRepository{OrderPaging: &repositories.OrderPaging{
		MaxSize: 100,
		CanOrderFields: map[string]repositories.Order{
			"id":          repositories.ASC | repositories.DESC,
			"name":        repositories.ASC | repositories.DESC,
			"message":     repositories.ASC | repositories.DESC,
			"level":       repositories.ASC | repositories.DESC,
			"create_time": repositories.ASC | repositories.DESC,
		},
	}, db: db}
}

// logRepository is a "Repository"
// which manages the accounts using the memory data source (map).
type logRepository struct {
	*repositories.OrderPaging
	db *sql.DB
}

func (r *logRepository) InsertLog(log *models.Log) (err error) {
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

	result, err := tx.Exec(_INSERT_SQL, &log.Name, &log.Message, &log.Level)
	if err != nil {
		return
	}
	log.ID, err = result.LastInsertId()
	success = true
	return
}

func (r *logRepository) FindLogByID(id int64) (log models.Log, err error) {
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
		&log.ID, &log.Name, &log.Message, &log.Level, &log.CreateTime)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *logRepository) FindLogByName(name string, page uint32, size uint32,
	sorts []repositories.Sort) (logs []models.Log, err error) {
	logs = make([]models.Log, 0, 50)
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
	rows, err := tx.Query(_SELECT_BY_NAME_SQL+paSoStr, name)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result models.Log
		err = rows.Scan(&result.ID, &result.Name, &result.Message, &result.Level, &result.CreateTime)
		if err != nil {
			return
		}
		logs = append(logs, result)
	}
	success = true
	return
}

func (r *logRepository) FindAll(page uint32, size uint32, sorts []repositories.Sort) (logs []models.Log, err error) {
	logs = make([]models.Log, 0, 100)
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
		var result models.Log
		err = rows.Scan(&result.ID, &result.Name, &result.Message, &result.Level, &result.CreateTime)
		if err != nil {
			return
		}
		logs = append(logs, result)
	}
	success = true
	return
}

func (r *logRepository) CountAll() (count int64, err error) {
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

func (r *logRepository) CountByName(name string) (count int64, err error) {
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

	err = tx.QueryRow(_COUNT_BY_NAME_SQL).Scan(&count)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *logRepository) DeleteLogByName(name string) (err error) {
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

	_, err = tx.Exec(_DELETE_BY_NAME_SQL, name)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *logRepository) DeleteLogByID(id int64) (log models.Log, err error) {
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
		&log.ID, &log.Name, &log.Message, &log.Level, &log.CreateTime)
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

func (r *logRepository) DeleteAll() (err error) {
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

func (r *logRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
