package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var SEPARATOR = []byte("\r\n")
var END_BYTE = []byte("\r\n\r\n")
var ErrFoo = fmt.Errorf("start line is bad")

func isTokenValid(fieldKey []byte) bool {
	for _, ch := range fieldKey {
		isValid := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= 0 && ch <= 9 {
			isValid = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			isValid = true
		}
		if !isValid {
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	headerKeyVal := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(headerKeyVal) != 2 {
		return "", "", ErrFoo
	}
	name := headerKeyVal[0]
	value := bytes.TrimSpace(headerKeyVal[1])
	if bytes.HasSuffix(name, []byte(" ")) || bytes.HasPrefix(name, []byte(" ")) {
		return "", "", ErrFoo
	}
	return string(name), string(value), nil
}

func (h *Headers) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}

func (h *Headers) HasKey(key string) bool {
	_, ok := h.headers[strings.ToLower(key)]
	return ok

}

func (h *Headers) ForEach(cb func(key string, value string)) {
	for key, value := range h.headers {
		cb(key, value)
	}
}

func (h *Headers) Size() int {
	return len(h.headers)
}
func (h *Headers) Set(key string, value string) {
	val, ok := h.headers[strings.ToLower(key)]
	if !ok {
		h.headers[strings.ToLower(key)] = value
	} else {
		h.headers[strings.ToLower(key)] = fmt.Sprintf("%s,%s", val, value)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], SEPARATOR)
		if idx == -1 {
			break
		}
		if idx == 0 {
			done = true
			read += len(SEPARATOR)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, fmt.Errorf("unable to parse header")
		}
		if !isTokenValid([]byte(name)) {
			return 0, false, fmt.Errorf("invalid header field key")
		}
		h.Set(name, value)
		read += idx + len(SEPARATOR)
	}

	return read, done, nil
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}
