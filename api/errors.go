package api

import (
	"fmt"
)

type MethodNotFoundError struct {
	Method string
}

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("the method %s does not exist/is not available", e.Method)
}

type InvalidRequestError struct{ Message string }

func (e *InvalidRequestError) Error() string { return e.Message }

type InvalidMessageError struct{ Message string }

func (e *InvalidMessageError) Error() string { return e.Message }

type InvalidParamsError struct{ Message string }

func (e *InvalidParamsError) Error() string { return e.Message }

type CallbackError struct{ Message string }

func (e *CallbackError) Error() string { return e.Message }

type ShutdownError struct{}

func (e *ShutdownError) Error() string { return "server is shutting down" }
