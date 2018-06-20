package types

import (
	"time"
)

// ProcessStatistics is a wrapper with the process current Statistics info.
type ProcessStatistics struct {
	Restarts  int           `json:"restart_count"`
	StartTime time.Time     `json:"start_time"`
	UpTime    time.Duration `json:"uptime"`
}

func (p *ProcessStatistics) AddRestart() {
	p.Restarts++
}

func (p *ProcessStatistics) InitUpTime() {
	p.StartTime = time.Now()
}

func (p *ProcessStatistics) SetUpTime() {
	p.UpTime = time.Since(p.StartTime)
}

func (p *ProcessStatistics) ResetUpTime() {
	p.UpTime = time.Duration(0)
}
