package types

import (
	"io"
	"os"

	"github.com/zhsyourai/URCF-engine/models"
)

type ProcessStatus int

const (
	Running ProcessStatus = iota
)

type Process struct {
	models.ProcessParam
	Pid        int
	PidFile    string
	StdIn      io.Writer
	StdOut     io.Reader
	StdErr     io.Reader
	DataOut    io.Reader
	KeepAlive  bool
	Statistics ProcessStatistics
	Status     ProcessStatus
	Process    *os.Process
}
