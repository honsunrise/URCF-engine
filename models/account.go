package models

import "time"

type Account struct {
	ID         string
	CreateDate time.Time
	Password   []byte
	Role       []string
	Enabled    bool
	UpdateDate time.Time
}
