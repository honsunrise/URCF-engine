package models

import (
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
