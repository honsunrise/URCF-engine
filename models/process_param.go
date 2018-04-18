package models

import (
	"database/sql/driver"
	"strings"
	"go/types"
	"errors"
	"fmt"
)

type ProcessOption int

const (
	None                      = 0
	AutoRestart ProcessOption = 1 << iota
	HookLog
)

func (option ProcessOption) String() (ret string) {
	if option == None {
		return "none"
	} else {
		if option & AutoRestart == AutoRestart {
			ret += ",autoRestart"
		}
		if option & HookLog == HookLog {
			ret += ",hookLog"
		}
		return strings.TrimPrefix(ret, ",")
	}
}

func ParseOption(option string) (ret ProcessOption, err error) {
	options := strings.Split(option, ",")
	for _, opt := range options {
		switch opt {
		case "autoRestart":
			ret |= AutoRestart
		case "hookLog":
			ret |= HookLog
		case "none":
			return None, nil
		default:
			return None, fmt.Errorf("not a valid ProcessOption: %q", opt)
		}
	}
	return ret, nil
}

func (option ProcessOption) Value() (driver.Value, error) {
	return option.String(), nil
}

func (option *ProcessOption) Scan(value interface{}) (err error) {
	switch value.(type) {
	case string:
		*option, err = ParseOption(value.(string))
	case []byte:
		*option, err = ParseOption(string(value.([]byte)))
	case types.Nil:
		*option = None
	default:
		return errors.New("failed to scan ProcessOption")
	}
	return nil
}

type Args []string

func (args Args) Value() (driver.Value, error) {
	return strings.Join(args, " "), nil
}

func (args *Args) Scan(value interface{}) error {
	switch value.(type) {
	case string:
		*args = strings.Split(value.(string), " ")
	case []byte:
		*args = strings.Split(string(value.([]byte)), " ")
	case types.Nil:
		*args = []string{}
	default:
		return errors.New("failed to scan Args")
	}
	return nil
}

type Env map[string]string

type ProcessParam struct {
	Name    string
	Cmd     string
	Args    Args
	WorkDir string
	Env     Env
	Option  ProcessOption
}
