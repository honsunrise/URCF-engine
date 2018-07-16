package server

import (
	"fmt"
)

func GetErrorCode(err error) int {
	switch err.(type) {
	case *methodNotFoundError:
		return -32601
	case *invalidRequestError:
		return -32600
	case *invalidMessageError:
		return -32700
	case *invalidParamsError:
		return -32602
	case *callbackError:
		return -32000
	case *shutdownError:
		return -32000
	}
	return 0
}

type methodNotFoundError struct {
	service string
	method  string
}

func (e *methodNotFoundError) Error() string {
	return fmt.Sprintf("The method %s%s%s does not exist/is not available",
		e.service, ServiceMethodSeparator, e.method)
}

type invalidRequestError struct{ message string }

func (e *invalidRequestError) Error() string { return e.message }

type invalidMessageError struct{ message string }

func (e *invalidMessageError) Error() string { return e.message }

type invalidParamsError struct{ message string }

func (e *invalidParamsError) Error() string { return e.message }

type callbackError struct{ message string }

func (e *callbackError) Error() string { return e.message }

type shutdownError struct{}

func (e *shutdownError) Error() string { return "server is shutting down" }
