package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

type requestStatus int

const (
	requestStateInitialized    requestStatus = iota // 0
	requestStateDone                                // 1
	requestStateParsingHeaders                      //2
)

type Request struct {
	Headers     headers.Headers
	RequestLine RequestLine
	State       requestStatus
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		// still consuming
		if n == 0 {
			return 0, nil
		}
		r.State = requestStateParsingHeaders
		r.RequestLine = *requestLine
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if !done {
			return n, nil
		}
		if done {
			r.State = requestStateDone
		}
		return n, nil

	case requestStateDone:
		return 0, fmt.Errorf("should not be parsing anymore, we are done")
	default:
		return 0, fmt.Errorf("unknown error state")
	}

}

// String() returns the RequestLine as a string
func (r *Request) String() string {
	return fmt.Sprintf("Request line: \n- Method: %s\n- Target: %s\n- Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	request := &Request{
		State:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}

	for request.State != requestStateDone {
		// Grow buffer
		if len(buf) == readToIndex {
			newSlice := make([]byte, len(buf)*2, len(buf)*2)
			copy(newSlice, buf)
			buf = newSlice
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				if request.State != requestStateDone {
					return request, fmt.Errorf("incomplete request: unexpected EOF")
				}
				break
			}
			return nil, fmt.Errorf("error reading from reader: %v", err)
		}
		// Update where our buffer is at now
		readToIndex += n
		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("failed to parse request: %v", err)
		}
		copy(buf, buf[bytesParsed:readToIndex])
		readToIndex -= bytesParsed
	}

	return request, nil

}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := parseRequestLineString(string(requestLineText))
	if err != nil {
		return nil, idx + 1, err
	}

	return requestLine, idx + 2, nil

}

func parseRequestLineString(reqLine string) (*RequestLine, error) {
	splits := strings.Split(string(reqLine), "\r")

	splits = strings.Split(splits[0], " ")

	if len(splits) != 3 {
		return nil, fmt.Errorf("invalid request line, not the correct number of parts. parts: %d", len(splits))
	}

	for _, c := range splits[0] {
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return nil, fmt.Errorf("invalid method: %s", splits[0])
		}
	}

	if splits[2] != "HTTP/1.1" {
		return nil, fmt.Errorf("request is not HTTP/1.1")
	}

	request := &RequestLine{
		Method:        splits[0],
		RequestTarget: splits[1],
		HttpVersion:   strings.Split(splits[2], "/")[1],
	}

	return request, nil
}
