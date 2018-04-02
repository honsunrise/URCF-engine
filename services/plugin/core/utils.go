package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
)

func GetRandomListenerAddr(forceTcp bool) (net.Addr, error) {
	lis, err := Listener(forceTcp)
	if err != nil {
		return nil, err
	}
	defer lis.Close()

	return lis.Addr(), nil
}

func Listener(forceTcp bool) (net.Listener, error) {
	if runtime.GOOS == "windows" || forceTcp {
		return listener_tcp()
	}

	return listener_unix()
}

func listener_tcp() (net.Listener, error) {
	address := fmt.Sprintf("127.0.0.1:0")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, errors.New("couldn't bind plugin TCP listener")
	}
	return listener, nil
}

func listener_unix() (net.Listener, error) {
	tf, err := ioutil.TempFile("", "urcf-plugin-")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	return &cleanUpListener{
		Listener: l,
		Path:     path,
	}, nil
}

type cleanUpListener struct {
	net.Listener
	Path string
}

func (l *cleanUpListener) Close() error {
	if err := l.Listener.Close(); err != nil {
		return err
	}

	return os.Remove(l.Path)
}
