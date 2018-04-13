package models

import (
	"time"
	"database/sql/driver"
	"errors"
	"strings"
	"go/types"
)

type Roles []string

func (roles Roles) Value() (driver.Value, error) {
	return strings.Join(roles, ","), nil
}

func (roles *Roles) Scan(value interface{}) error {
	switch value.(type) {
	case string:
		*roles = strings.Split(value.(string), ",")
	case []byte:
		*roles = strings.Split(string(value.([]byte)), ",")
	case types.Nil:
		*roles = []string{}
	default:
		return errors.New("failed to scan Roles")
	}
	return nil
}

type Account struct {
	ID         int64
	Username   string
	CreateDate time.Time
	Password   []byte
	Roles      Roles
	Enabled    bool
	UpdateDate time.Time
}
