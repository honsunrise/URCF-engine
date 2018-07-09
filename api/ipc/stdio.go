package ipc

import (
	"net"
	"os"
	"time"
)

type StdIOConn struct{}

func (io StdIOConn) Read(b []byte) (n int, err error) {
	return os.Stdin.Read(b)
}

func (io StdIOConn) Write(b []byte) (n int, err error) {
	return os.Stdout.Write(b)
}

func (io StdIOConn) Close() error {
	return nil
}

func (io StdIOConn) LocalAddr() net.Addr {
	return &net.UnixAddr{Name: "stdio", Net: "stdio"}
}

func (io StdIOConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{Name: "stdio", Net: "stdio"}
}

func (io StdIOConn) SetDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "stdio", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (io StdIOConn) SetReadDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "stdio", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (io StdIOConn) SetWriteDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "stdio", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}
