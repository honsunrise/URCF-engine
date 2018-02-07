package processes

import (
	"io"
	"os"

	"github.com/zhsyourai/URCF-engine/models"
)

type ProcessStatus int

const (
	Running ProcessStatus = iota,
)

type Process struct {
	models.ProcessParam
	Pid        int
	PidFile    string
	StdIn      io.ReadWriter
	StdOut     io.ReadWriter
	StdErr     io.ReadWriter
	KeepAlive  bool
	Statistics ProcessStatistics
	Status     ProcessStatus
	Process    *os.Process
}
