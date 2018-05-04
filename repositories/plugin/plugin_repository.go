package plugin

import (
	"io"
	"log"
	"reflect"

	"database/sql"
	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"os"
	"path"
)

const (
	_CREATE_TABLE_SQL_ = `CREATE TABLE IF NOT EXISTS plugins (
			name TEXT PRIMARY KEY,
			desc TEXT,
			maintainer TEXT NOT NULL,
			homepage TEXT NOT NULL,
			version TEXT NOT NULL,
			enter_point TEXT NOT NULL,
			enable BOOLEAN NOT NULL,
			install_dir TEXT NOT NULL,
			webs_dir TEXT NOT NULL,
			cover_file TEXT NOT NULL,
			install_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL
		)`

	_INSERT_SQL = `INSERT INTO plugins(name, desc, maintainer, homepage, version, enter_point, enable, install_dir, webs_dir, cover_file, install_time, update_time)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_SELECT_ALL_SQL = `SELECT * FROM plugins`

	_SELECT_BY_NAME_SQL = `SELECT * FROM plugins WHERE name = ?`

	_COUNT_ALL_SQL = `SELECT COUNT(*) as count FROM plugins`

	_DELETE_BY_NAME_SQL = `DELETE FROM plugins WHERE name = ?`

	_UPDATE_BY_NAME_SQL = `UPDATE plugins SET enable = ?, update_time = CURRENT_TIMESTAMP WHERE name = ?`
)

// Repository handles the basic operations of a plugin entity/model.
// It's an interface in order to be testable, i.e a memory plugin repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertPlugin(plugin *models.Plugin) error
	FindPluginByName(name string) (models.Plugin, error)
	FindAll(page uint32, size uint32, sorts []repositories.Sort) ([]models.Plugin, error)
	CountAll() (int64, error)
	DeletePluginByName(name string) (models.Plugin, error)
	UpdatePluginByName(name string, fields map[string]interface{}) (plugin models.Plugin, err error)
}

// NewPluginRepository returns a new plugin memory-based repository,
// the one and only repository type in our example.
func NewPluginRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbPath := confServ.Get().Sys.DatabasePath
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0770)
	}
	dbFile := path.Join(dbPath, "Plugin.db")

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(_CREATE_TABLE_SQL_)
	if err != nil {
		log.Fatal(err)
	}
	return &pluginRepository{OrderPaging: &repositories.OrderPaging{
		MaxSize: 100,
		CanOrderFields: map[string]repositories.Order{
			"enable": repositories.ASC | repositories.DESC,
		},
	}, db: db}
}

// pluginRepository is a "Repository"
// which manages the plugins using the memory data source (map).
type pluginRepository struct {
	*repositories.OrderPaging
	db *sql.DB
}

func (r *pluginRepository) InsertPlugin(inputPlugin *models.Plugin) (err error) {
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

	_, err = tx.Exec(_INSERT_SQL, &inputPlugin.Name, &inputPlugin.Desc, &inputPlugin.Maintainer, &inputPlugin.Homepage,
		&inputPlugin.Version, &inputPlugin.EnterPoint, &inputPlugin.Enable, &inputPlugin.InstallDir,
		&inputPlugin.WebsDir, &inputPlugin.CoverFile)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *pluginRepository) FindPluginByName(name string) (plugin models.Plugin, err error) {
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

	err = tx.QueryRow(_SELECT_BY_NAME_SQL, name).Scan(
		&plugin.Name, &plugin.Desc, &plugin.Maintainer, &plugin.Homepage,
		&plugin.Version, &plugin.EnterPoint, &plugin.Enable, &plugin.InstallDir,
		&plugin.WebsDir, &plugin.CoverFile, &plugin.InstallTime, &plugin.UpdateTime)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *pluginRepository) FindAll(page uint32, size uint32, sorts []repositories.Sort) (plugins []models.Plugin,
	err error) {
	plugins = make([]models.Plugin, 0, 100)
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
		var plugin models.Plugin
		err = rows.Scan(&plugin.Name, &plugin.Desc, &plugin.Maintainer, &plugin.Homepage,
			&plugin.Version, &plugin.EnterPoint, &plugin.Enable, &plugin.InstallDir,
			&plugin.WebsDir, &plugin.CoverFile, &plugin.InstallTime, &plugin.UpdateTime)
		if err != nil {
			return
		}
		plugins = append(plugins, plugin)
	}
	success = true
	return
}

func (r *pluginRepository) CountAll() (count int64, err error) {
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

func (r *pluginRepository) DeletePluginByName(name string) (plugin models.Plugin, err error) {
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
	err = tx.QueryRow(_SELECT_BY_NAME_SQL, name).Scan(
		&plugin.Name, &plugin.Desc, &plugin.Maintainer, &plugin.Homepage,
		&plugin.Version, &plugin.EnterPoint, &plugin.Enable, &plugin.InstallDir,
		&plugin.WebsDir, &plugin.CoverFile, &plugin.InstallTime, &plugin.UpdateTime)
	if err != nil {
		return
	}

	_, err = tx.Exec(_DELETE_BY_NAME_SQL, name)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *pluginRepository) UpdatePluginByName(name string,
	fields map[string]interface{}) (plugin models.Plugin, err error) {
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
	err = tx.QueryRow(_SELECT_BY_NAME_SQL, name).Scan(
		&plugin.Name, &plugin.Desc, &plugin.Maintainer, &plugin.Homepage,
		&plugin.Version, &plugin.EnterPoint, &plugin.Enable, &plugin.InstallDir,
		&plugin.WebsDir, &plugin.CoverFile, &plugin.InstallTime, &plugin.UpdateTime)
	if err != nil {
		return
	}

	s := reflect.ValueOf(&plugin).Elem()
	for k, v := range fields {
		field := s.FieldByName(k)
		if field.IsValid() {
			field.Set(reflect.ValueOf(v))
		} else {
			err = errors.New(fmt.Sprintf("field %s not exist", k))
			return
		}
	}

	_, err = tx.Exec(_UPDATE_BY_NAME_SQL, plugin.Name, plugin.Enable, name)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *pluginRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
