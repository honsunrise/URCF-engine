package processes

import (
	"time"
)

// ProcessStatistics is a wrapper with the process current Statistics info.
type ProcessStatistics struct {
	Restarts  int
	StartTime time.Time
	UpTime    time.Duration
}

func (p *ProcessStatistics) AddRestart() {
	p.Restarts++
}

func (p *ProcessStatistics) InitUpTime() {
	p.StartTime = time.Now()
}

func (p *ProcessStatistics) SetUpTime() {
	p.UpTime = time.Now().Sub(p.StartTime)
}

func (p *ProcessStatistics) ResetUptime() {
	p.UpTime = time.Duration(0)
}
