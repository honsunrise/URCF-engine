package processes

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"io"
)

type Service interface {
	Start() error
	Restart() error
	Stop() error
	Kill() error
	Watch() error
	IsAlive() bool
	GetPid() int
	GetStatus() int
	GetStdIn() io.ReadWriter
	GetStdOut() io.ReadWriter
	GetStdErr() io.ReadWriter
	GetWorkDir() string
}

// Process is a os.Process wrapper with Statistics and more info that will be used on Master to maintain
// the process health.
type Process struct {
	Name       string
	Pid        int
	Cmd        string
	Args       []string
	Path       string
	PidFile    string
	StdIn      io.ReadWriter
	StdOut     io.ReadWriter
	StdErr     io.ReadWriter
	KeepAlive  bool
	Statistics *ProcessStatistics
	process    *os.Process
}

func (p *Process) Start() error {
	outFile, err := utils.GetFile(proc.Outfile)
	if err != nil {
		return err
	}
	errFile, err := utils.GetFile(proc.Errfile)
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	procAtr := &os.ProcAttr{
		Dir: wd,
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			outFile,
			errFile,
		},
	}
	args := append([]string{proc.Name}, proc.Args...)
	process, err := os.StartProcess(proc.Cmd, args, procAtr)
	if err != nil {
		return err
	}
	proc.process = process
	proc.Pid = proc.process.Pid
	err = utils.WriteFile(proc.Pidfile, []byte(strconv.Itoa(proc.process.Pid)))
	if err != nil {
		return err
	}
	proc.Status.InitUptime()
	proc.Status.SetStatus("started")
	return nil
}

func (p *Process) Restart() error {
	if proc.IsAlive() {
		err := proc.GracefullyStop()
		if err != nil {
			return err
		}
	}
	return proc.Start()
}

func (p *Process) Stop() error    {
	if proc.process != nil {
		err := proc.process.Signal(syscall.SIGTERM)
		proc.Status.SetStatus("asked to stop")
		return err
	}
	return errors.New("Process does not exist.")
}

func (p *Process) Kill() error {
	if proc.process != nil {
		err := proc.process.Signal(syscall.SIGKILL)
		proc.Status.SetStatus("stopped")
		proc.release()
		return err
	}
	return errors.New("Process does not exist.")
}

func (p *Process) Watch() error             {}
func (p *Process) IsAlive() bool            {}
func (p *Process) GetPid() int              {}
func (p *Process) GetStatus() int           {}
func (p *Process) GetStdIn() io.ReadWriter  {}
func (p *Process) GetStdOut() io.ReadWriter {}
func (p *Process) GetStdErr() io.ReadWriter {}
func (p *Process) GetWorkDir() string       {}

// Delete will delete everything created by this process, including the out, err and pid file.
// Returns an error in case there's any.
func (proc *Process) Delete() error {
	proc.release()
	err := utils.DeleteFile(proc.Outfile)
	if err != nil {
		return err
	}
	err = utils.DeleteFile(proc.Errfile)
	if err != nil {
		return err
	}
	return os.RemoveAll(proc.Path)
}

// IsAlive will check if the process is alive or not.
// Returns true if the process is alive or false otherwise.
func (proc *Process) IsAlive() bool {
	p, err := os.FindProcess(proc.Pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// Watch will stop execution and wait until the process change its state. Usually changing state, means that the process died.
// Returns a tuple with the new process state and an error in case there's any.
func (proc *Process) Watch() (*os.ProcessState, error) {
	return proc.process.Wait()
}

// Will release the process and remove its PID file
func (proc *Process) release() {
	if proc.process != nil {
		proc.process.Release()
	}
	utils.DeleteFile(proc.Pidfile)
}

// NotifyStopped that process was stopped so we can set its PID to -1
func (proc *Process) NotifyStopped() {
	proc.Pid = -1
}

// AddRestart is add one restart to proc status
func (proc *Process) AddRestart() {
	proc.Statistics.AddRestart()
}

// GetPid will return proc current PID
func (proc *Process) GetPid() int {
	return proc.Pid
}

// GetOutFile will return proc out file
func (proc *Process) GetOutFile() string {
	return proc.Outfile
}

// GetErrFile will return proc error file
func (proc *Process) GetErrFile() string {
	return proc.Errfile
}

// GetPidFile will return proc pid file
func (proc *Process) GetPidFile() string {
	return proc.Pidfile
}

// GetPath will return proc path
func (proc *Process) GetPath() string {
	return proc.Path
}

// GetStatus will return proc current status
func (proc *Process) GetStatus() *ProcessStatistics {
	if !proc.IsAlive() {
		proc.ResetUpTime()
	} else {
		// update uptime
		proc.SetUptime()
	}
	// update cpu and memory
	proc.SetSysInfo()

	return proc.Statistics
}

// SetStatus will set proc status
func (proc *Process) SetStatus(status string) {
	proc.Statistics.SetStatus(status)
}

// SetUpTime will set UpTime
func (proc *Process) SetUptime() {
	proc.Statistics.SetUpTime()
}

// ResetUpTime will set UpTime
func (proc *Process) ResetUpTime() {
	proc.Statistics.ResetUptime()
}

// SetSysInfo will get current proc cpu and memory usage
func (proc *Process) SetSysInfo() {
	proc.Statistics.SetSysInfo(proc.process.Pid)
}

// Identifier is that will be used by watcher to keep track of its processes
func (proc *Process) Identifier() string {
	return proc.Name
}

// ShouldKeepAlive will returns true if the process should be kept alive or not
func (proc *Process) ShouldKeepAlive() bool {
	return proc.KeepAlive
}

// GetName will return current proc name
func (proc *Process) GetName() string {
	return proc.Name
}
