package models

import (
	"time"
	"github.com/zhsyourai/URCF-engine/utils"
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
