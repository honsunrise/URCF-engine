package models

type ProcessOption int

const (
	AutoRestart ProcessOption = iota
)

type ProcessParam struct {
	Name    string
	Cmd     string
	Args    []string
	WorkDir string
	Env     map[string]string
	Option  ProcessOption
}
