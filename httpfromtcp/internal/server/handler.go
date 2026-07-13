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

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.Code)

	messageBytes := []byte(he.Message)
	defaultHeaders := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, defaultHeaders)
	w.Write(messageBytes)
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func NewHandlerError(code response.StatusCode, msg string) *HandlerError {
	return &HandlerError{
		Code:    code,
		Message: msg,
	}
}
