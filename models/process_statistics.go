package models

import (
	"sync"
	"time"
)

// ProcessStatistics is a wrapper with the process current Statistics info.
type ProcessStatistics struct {
	lock         sync.Mutex
	Restarts     uint32    `json:"restart_count"`
	Stops        uint32    `json:"stop_count"`
	StartUpTime  time.Time `json:"startup_time"`
	LastStopTime time.Time `json:"laststop_time"`
}

func (p *ProcessStatistics) AddRestart() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Restarts++
}

func (p *ProcessStatistics) AddStop() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Stops++
}

func (p *ProcessStatistics) InitStartUpTime() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.StartUpTime = time.Now()
}

func (p *ProcessStatistics) SetLastStopTime() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.LastStopTime = time.Now()
}
