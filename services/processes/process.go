package processes

import (
	"io"
	"os"
)

type ProcessStatus int

const (
	Running ProcessStatus = iota,
)

type Process struct {
	Name       string
	Cmd        string
	Args       []string
	WorkDir    string
	Env        map[string]string
	Pid        int
	PidFile    string
	StdIn      io.ReadWriter
	StdOut     io.ReadWriter
	StdErr     io.ReadWriter
	KeepAlive  bool
	Statistics ProcessStatistics
	Status     ProcessStatus
	process    *os.Process
}
