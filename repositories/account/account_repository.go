package account

import (
	"io"
	"log"
	"reflect"

	"errors"
	"fmt"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"path"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
)

const (
	_CREATE_TABLE_SQL_ = `CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password BLOB NOT NULL,
			enable BOOLEAN NOT NULL,
			roles TEXT NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL
		);`

	_INSERT_USER_SQL = `INSERT INTO accounts(username, password, roles, enable, create_time, update_time)
			VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);`

	_SELECT_ALL_SQL = `SELECT * FROM accounts;`

	_SELECT_BY_USERNAME_SQL = `SELECT * FROM accounts WHERE username = ?;`

	_DELETE_BY_USERNAME_SQL = `DELETE FROM accounts WHERE username = ?;`

	_UPDATE_BY_USERNAME_SQL = `UPDATE accounts SET username = ?, password = ?, roles = ?, enable = ?, update_time = CURRENT_TIMESTAMP WHERE username = ?;`
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertAccount(models.Account) error
	FindAccountByUsername(username string) (models.Account, error)
	FindAll() ([]models.Account, error)
	DeleteAccountByUsername(username string) (models.Account, error)
	UpdateAccountByUsername(username string, account map[string]interface{}) (models.Account, error)
}

// NewAccountRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewAccountRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbFile := path.Join(confServ.Get().Sys.WorkPath, "database", "Account.db")

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(_CREATE_TABLE_SQL_)
	if err != nil {
		log.Fatal(err)
	}
	return &accountRepository{db}
}

// accountRepository is a "Repository"
// which manages the accounts using the memory data source (map).
type accountRepository struct {
	db *sql.DB
}

func (r *accountRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *accountRepository) InsertAccount(account models.Account) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			err = tx.Rollback()
		} else {
			if err = tx.Commit(); err != nil {
				err = tx.Rollback()
				return
			}
		}
	}()

	_, err = tx.Exec(_INSERT_USER_SQL, &account.Username, &account.Password, &account.Roles, &account.Enabled)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *accountRepository) FindAccountByUsername(username string) (account models.Account, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			err = tx.Rollback()
		} else {
			if err = tx.Commit(); err != nil {
				err = tx.Rollback()
				return
			}
		}
	}()

	err = tx.QueryRow(_SELECT_BY_USERNAME_SQL, username).Scan(
		&account.ID, &account.Username, &account.Password, &account.Roles, &account.Enabled,
		&account.CreateDate, &account.UpdateDate)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *accountRepository) FindAll() (accounts []models.Account, err error) {
	accounts = make([]models.Account, 0, 10)
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			err = tx.Rollback()
		} else {
			if err = tx.Commit(); err != nil {
				err = tx.Rollback()
				return
			}
		}
	}()

	rows, err := tx.Query(_SELECT_ALL_SQL)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var account models.Account
		err = rows.Scan(&account.ID, &account.Username, &account.Password, &account.Roles, &account.Enabled,
			&account.CreateDate, &account.UpdateDate)
		if err != nil {
			return
		}
		accounts = append(accounts, account)
	}
	success = true
	return
}

func (r *accountRepository) DeleteAccountByUsername(username string) (account models.Account, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			err = tx.Rollback()
		} else {
			if err = tx.Commit(); err != nil {
				err = tx.Rollback()
				return
			}
		}
	}()
	err = tx.QueryRow(_SELECT_BY_USERNAME_SQL, username).Scan(
		&account.ID, &account.Username, &account.Password, &account.Roles, &account.Enabled,
		&account.CreateDate, &account.UpdateDate)
	if err != nil {
		return
	}

	_, err = tx.Exec(_DELETE_BY_USERNAME_SQL, username)
	if err != nil {
		return
	}
	success = true
	return
}

func (r *accountRepository) UpdateAccountByUsername(username string,
	fields map[string]interface{}) (account models.Account, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	success := false
	defer func() {
		if !success {
			err = tx.Rollback()
		} else {
			if err = tx.Commit(); err != nil {
				err = tx.Rollback()
				return
			}
		}
	}()
	err = tx.QueryRow(_SELECT_BY_USERNAME_SQL, username).Scan(
		&account.ID, &account.Username, &account.Password, &account.Roles, &account.Enabled,
		&account.CreateDate, &account.UpdateDate)
	if err != nil {
		return
	}

	s := reflect.ValueOf(&account).Elem()
	for k, v := range fields {
		field := s.FieldByName(k)
		if field.IsValid() {
			field.Set(reflect.ValueOf(v))
		} else {
			err = errors.New(fmt.Sprintf("field %s not exist", k))
			return
		}
	}

	_, err = tx.Exec(_UPDATE_BY_USERNAME_SQL, account.Username, account.Password,
		account.Roles, account.Enabled, username)
	if err != nil {
		return
	}
	success = true
	return
}
