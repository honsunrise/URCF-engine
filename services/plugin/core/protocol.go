package core

import "github.com/zhsyourai/URCF-engine/utils/async"

type CommandProtocol interface {
	Command(name string, params ...string) async.AsyncRet
	GetHelp(name string) async.AsyncRet
	ListCommand() async.AsyncRet
}
