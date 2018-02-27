package models

import "time"

type Config struct {
	Key        string
	Value      interface{}
	CreateDate time.Time
	UpdateDate time.Time
	Scope      string
	Expires    time.Duration
}
