package protocol

type CommandProtocol interface {
	Command(name string, params ...string) (string, error)
	GetHelp(name string) (string, error)
	ListCommand() ([]string, error)
}
