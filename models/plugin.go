package models

import (
	"github.com/zhsyourai/URCF-engine/utils"
	"time"
)

type Plugin struct {
	Name        string
	Version     utils.SemanticVersion
	EnterPoint  string
	Enable      bool
	InstallDir  string
	InstallTime time.Time
	UpdateTime  time.Time
}
