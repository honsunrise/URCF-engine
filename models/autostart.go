package models

import (
	"time"
)

type AutoStart struct {
	ID         string
	CreateDate time.Time
	Priority   int32
	StartDelay int32
	StopDelay  int32
	Enable     bool
	Parallel   bool
	ProcessParam
}

type ByPriority []AutoStart

func (a ByPriority) Len() int {
	return len(a)
}

func (a ByPriority) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByPriority) Less(i, j int) bool {
	return a[i].Priority < a[j].Priority
}
