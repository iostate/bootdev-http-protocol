package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes: %v", err)
	}
	reqLine, err := parseRequestLine(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get request line: %v", err)
	}

	return &Request{RequestLine: *reqLine}, nil

}

func parseRequestLine(bytes []byte) (*RequestLine, error) {
	splits := strings.Split(string(bytes), "\r")

	splits = strings.Split(splits[0], " ")

	if len(splits) > 3 {
		return nil, fmt.Errorf("invalid request line, has too many parts. parts: %d", len(splits))
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
