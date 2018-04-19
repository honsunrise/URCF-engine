package models

import "time"

type Config struct {
	Key        string
	Value      interface{}
	CreateTime time.Time
	UpdateTime time.Time
	Expires    time.Duration
}
