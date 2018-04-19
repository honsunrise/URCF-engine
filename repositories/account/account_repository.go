package account

import (
	"io"
	"log"
	"reflect"

	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/services/global_configuration"
	"os"
	"path"
)

var (
	ErrAccountExist = errors.New("account exist")
)

const (
	_CREATE_TABLE_SQL_ = `CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password BLOB NOT NULL,
			roles TEXT NOT NULL,
			enable BOOLEAN NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL
		)`

	_INSERT_USER_SQL = `INSERT INTO accounts(username, password, roles, enable, create_time, update_time)
			VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_EXIST_BY_USERNAME_SQL = `SELECT EXISTS(SELECT id FROM accounts WHERE username = ?)`

	_SELECT_ALL_SQL = `SELECT * FROM accounts`

	_SELECT_BY_USERNAME_SQL = `SELECT * FROM accounts WHERE username = ?`

	_DELETE_BY_USERNAME_SQL = `DELETE FROM accounts WHERE username = ?`

	_UPDATE_BY_USERNAME_SQL = `UPDATE accounts SET username = ?, password = ?, roles = ?, enable = ?, update_time = CURRENT_TIMESTAMP WHERE username = ?`
)

// Repository handles the basic operations of a account entity/model.
// It's an interface in order to be testable, i.e a memory account repository or
// a connected to an sql database.
type Repository interface {
	io.Closer
	InsertAccount(account *models.Account) error
	FindAccountByUsername(username string) (models.Account, error)
	FindAll() ([]models.Account, error)
	DeleteAccountByUsername(username string) (models.Account, error)
	UpdateAccountByUsername(username string, account map[string]interface{}) (models.Account, error)
}

// NewAccountRepository returns a new account memory-based repository,
// the one and only repository type in our example.
func NewAccountRepository() Repository {
	confServ := global_configuration.GetGlobalConfig()
	dbPath := path.Join(confServ.Get().Sys.WorkPath, "database")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0770)
	}
	dbFile := path.Join(dbPath, "Account.db")

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

func (r *accountRepository) InsertAccount(account *models.Account) (err error) {
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

	var exists bool

	err = tx.QueryRow(_EXIST_BY_USERNAME_SQL, account.Username).Scan(&exists)
	if err != nil {
		return
	}

	if exists {
		err = ErrAccountExist
		return
	}

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
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
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
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
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
			if e := tx.Rollback(); e != nil {
				err = e
			}
		} else {
			err = tx.Commit()
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
