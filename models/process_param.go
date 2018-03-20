package models

type ProcessOption int

const (
	None = 0
	AutoRestart ProcessOption = 1 << iota
	HookLog
)

type ProcessParam struct {
	Name    string
	Cmd     string
	Args    []string
	WorkDir string
	Env     map[string]string
	Option  ProcessOption
}
