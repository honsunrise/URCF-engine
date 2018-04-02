package types

import (
	"io"
	"os"

	"github.com/zhsyourai/URCF-engine/models"
)

type ProcessStatus int

const (
	Prepare ProcessStatus = iota
	Running
	Exiting
	Exited
)

type Process struct {
	models.ProcessParam
	Pid        int
	PidFile    string
	StdIn      io.WriteCloser
	StdOut     io.ReadCloser
	StdErr     io.ReadCloser
	DataOut    io.ReadCloser
	KeepAlive  bool
	Statistics ProcessStatistics
	Status     ProcessStatus
	Process    *os.Process
	ExitChan   chan struct{}
}
