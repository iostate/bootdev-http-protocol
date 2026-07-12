package response

import (
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	switch statusCode {
	case Ok:
		reasonPhrase = "OK"
	case BadRequest:
		reasonPhrase = "Bad Request"
	case InternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}

	responseStr := "HTTP/1.1 " + strconv.Itoa(int(statusCode)) + " " + reasonPhrase + "\r\n"
	_, err := w.Write([]byte(responseStr))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
