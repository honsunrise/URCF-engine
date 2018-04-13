package models

import (
	"fmt"
	"strings"
	"time"
	"database/sql/driver"
	"go/types"
	"errors"
)

type Level uint32

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	}

	return "unknown"
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "err", "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid Level: %q", lvl)
}

func (level Level) Value() (driver.Value, error) {
	return level.String(), nil
}

func (level *Level) Scan(value interface{}) (err error) {
	switch value.(type) {
	case string:
		*level, err = ParseLevel(value.(string))
	case []byte:
		*level, err = ParseLevel(string(value.([]byte)))
	case types.Nil:
		*level = DebugLevel
	default:
		return errors.New("failed to scan Level")
	}
	return nil
}

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

type Log struct {
	ID         int64     `json:"id"`
	Message    string    `json:"message"`
	Name       string    `json:"name"`
	Level      Level     `json:"level"`
	CreateTime time.Time `json:"create_time"`
}
