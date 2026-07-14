package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {
	keyLower := strings.ToLower(key)
	return h[keyLower]
}

func (h Headers) Set(key, value string) string {
	keyLower := strings.ToLower(key)
	h[keyLower] = strings.ToLower(value)
	return h[keyLower]
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return len(crlf), true, nil
	}

	// idx > 0 branch = real header line found
	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := parts[0]
	if !bytes.Equal(parts[0], bytes.TrimSpace(parts[0])) {
		return 0, false, fmt.Errorf("failed to parse header key")
	}
	if len(key) == 0 {
		return 0, false, fmt.Errorf("invalid header key with length 0")
	}
	for _, c := range key {
		if !isValidHeaderKeyChar(c) {
			return 0, false, fmt.Errorf("failed to parse header key")
		}
	}

	keyStr := strings.ToLower(string(key))
	value := bytes.TrimSpace(parts[1])
	valueStr := strings.ToLower(string(value))

	// Check if key already exists
	if _, ok := h[keyStr]; ok {
		h[keyStr] += ", " + valueStr
	} else {
		h[keyStr] = valueStr
	}

	return idx + len(crlf), false, nil

}
func isValidHeaderKeyChar(c byte) bool {
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
		return true
	}
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}
