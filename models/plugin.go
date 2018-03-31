package models

import (
	"github.com/zhsyourai/URCF-engine/utils"
	"time"
)

type Plugin struct {
	ID          string
	Title       string
	Enabled     bool
	InstallDate time.Time
	Path        string
	WorkDir     string
	EnterPoint  []string
	Version     utils.SemanticVersion
}
