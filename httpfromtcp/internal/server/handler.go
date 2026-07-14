package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

func (he HandlerError) Write(w *response.Writer) {
	w.WriteStatusLine(he.Code)

	messageBytes := []byte(he.Message)
	defaultHeaders := response.GetDefaultHeaders(len(messageBytes))
	w.WriteHeaders(defaultHeaders)
	w.WriteBody(messageBytes)
}

type Handler func(w *response.Writer, req *request.Request)

func NewHandlerError(code response.StatusCode, msg string) *HandlerError {
	return &HandlerError{
		Code:    code,
		Message: msg,
	}
}
