package models

import (
	"time"
	"github.com/zhsyourai/URCF-engine/utils"
)

type Plugin struct {
	Name        string
	Version     utils.SemanticVersion
	EnterPoint  string
	Enable      bool
	InstallTime time.Time
	UpdateTime  time.Time
}
