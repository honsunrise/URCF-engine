package models

import "time"

type Account struct {
	ID         string
	CreateDate time.Time
	Password   string
	Role       []string
	Enabled    bool
	UpdateDate time.Time
}
