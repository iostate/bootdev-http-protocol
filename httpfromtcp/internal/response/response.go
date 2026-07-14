package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// State machine that ensures we write responses in order
type WriterState string

const (
	StateWriteStatusLine WriterState = "WriteStatusLine"
	StateWriteHeaders    WriterState = "WriteHeaders"
	StateWriteBody       WriterState = "WriteBody"
)

type Writer struct {
	writer      io.Writer
	writerState WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w, writerState: StateWriteStatusLine}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != StateWriteStatusLine {
		return fmt.Errorf("not in the correct state, state: %s", w.writerState)
	}
	var reasonPhrase string
	switch statusCode {
	case StatusOk:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}

	responseStr := "HTTP/1.1 " + strconv.Itoa(int(statusCode)) + " " + reasonPhrase + "\r\n"
	_, err := w.writer.Write([]byte(responseStr))
	if err != nil {
		return err
	}

	w.writerState = StateWriteHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"
	return headers
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != StateWriteBody {
		return 0, fmt.Errorf("not in the correct state, state: %s", w.writerState)
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != StateWriteBody {
		return 0, fmt.Errorf("not in the correct state, state: %s", w.writerState)
	}

	if len(p) == 0 {
		return 0, fmt.Errorf("empty chunk")
	}

	out := make([]byte, 0, len(p)+32)
	out = append(out, fmt.Sprintf("%x\r\n", len(p))...)
	out = append(out, p...)
	out = append(out, "\r\n"...)

	if _, err := w.writer.Write(out); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != StateWriteBody {
		return 0, fmt.Errorf("not in the correct state, state: %s", w.writerState)
	}
	n, err := w.writer.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != StateWriteHeaders {
		return fmt.Errorf("not in the correct state, state: %s", w.writerState)
	}
	for key, value := range headers {
		_, err := w.writer.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	if _, err := w.writer.Write([]byte("\r\n")); err != nil {
		return err
	}

	w.writerState = StateWriteBody
	return nil
}
