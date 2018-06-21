package types

import (
	"github.com/zhsyourai/URCF-engine/utils"
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

var processStrings = []utils.IntName{
	{0, "Prepare"},
	{1, "Running"},
	{2, "Exiting"},
	{3, "Exited"},
}

func (i ProcessStatus) String() string {
	return utils.StringName(uint32(i), processStrings, "processStatus.", false)
}
func (i ProcessStatus) GoString() string {
	return utils.StringName(uint32(i), processStrings, "processStatus.", true)
}
func (i ProcessStatus) MarshalText() ([]byte, error) {
	return []byte(utils.StringName(uint32(i), processStrings, "processStatus.", false)), nil
}

type Process struct {
	models.ProcessParam
	Pid        int               `json:"pid"`
	PidFile    string            `json:"pid_file"`
	StdIn      io.WriteCloser    `json:"-"`
	StdOut     io.ReadCloser     `json:"-"`
	StdErr     io.ReadCloser     `json:"-"`
	DataOut    io.ReadCloser     `json:"-"`
	Statistics ProcessStatistics `json:"statistics"`
	State      ProcessStatus     `json:"state,string"`
	Process    *os.Process       `json:"-"`
}
