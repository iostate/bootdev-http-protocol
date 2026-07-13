package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func NewHandlerError(code response.StatusCode, msg string) *HandlerError {
	return &HandlerError{
		Code:    code,
		Message: msg,
	}
}
