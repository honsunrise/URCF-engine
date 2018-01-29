package models

import "time"

type Plugin struct {
	ID          string
	Title       string
	Enabled     bool
	InstallDate time.Time
	Path        string
}
