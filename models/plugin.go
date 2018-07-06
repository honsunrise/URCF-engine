package models

import (
	"errors"
	"github.com/zhsyourai/URCF-engine/utils"
	"time"
)

type Plugin struct {
	Name        string                `json:"name"`
	Desc        string                `json:"desc"`
	Maintainer  string                `json:"maintainer"`
	Homepage    string                `json:"homepage"`
	Version     utils.SemanticVersion `json:"version"`
	EnterPoint  string                `json:"enter_point"`
	Enable      bool                  `json:"enable"`
	InstallDir  string                `json:"install_dir"`
	WebsDir     string                `json:"webs_dir"`
	CoverFile   string                `json:"cover"`
	InstallTime time.Time             `json:"install_time"`
	UpdateTime  time.Time             `json:"update_time"`
}

type Protocol int32

const (
	NoneProtocol Protocol = iota
	JsonRPCProtocol
)

var protocolStrings = []utils.IntName{
	{0, "NoneProtocol"},
	{1, "JsonRPCProtocol"},
}

func (i Protocol) String() string {
	return utils.StringName(uint32(i), protocolStrings, "plugin.", false)
}

func (i Protocol) GoString() string {
	return utils.StringName(uint32(i), protocolStrings, "plugin.", true)
}

func (i Protocol) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Protocol) UnmarshalText(b []byte) error {
	*i = JsonRPCProtocol
	switch string(b) {
	case "NoneProtocol":
		*i = NoneProtocol
	case "JsonRPCProtocol":
		*i = JsonRPCProtocol
	default:
		return errors.New(string(b) + " not correct error")
	}
	return nil
}

type Protocols []Protocol

func (ps Protocols) String() string {
	var ret string
	for i, p := range ps {
		if i == 0 {
			ret = p.String()
		} else {
			ret += "," + p.String()
		}
	}
	return ret
}

func (ps Protocols) Exist(item Protocol) bool {
	for _, p := range ps {
		if p == item {
			return true
		}
	}
	return false
}
