package request

import (
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var ERROR_BAD_START_LINE = fmt.Errorf("start line is bad")
var SEPARATOR = "\r\n"

func parseRequestLine(httpRequest string) (*RequestLine, string, error) {

	idx := strings.Index(httpRequest, SEPARATOR)
	if idx == -1 {
		return nil, httpRequest, ERROR_BAD_START_LINE
	}
	startString := httpRequest[:idx]
	restOfTheString := httpRequest[idx+len(SEPARATOR):]

	req1 := strings.Split(startString, " ")
	if len(req1) != 3 {
		return nil, restOfTheString, ERROR_BAD_START_LINE
	}

	versionPart := strings.Split(req1[2], "/")
	if len(versionPart) != 2 || versionPart[0] != "HTTP" {
		return nil, restOfTheString, ERROR_BAD_START_LINE
	}

	r := &RequestLine{
		Method:        req1[0],
		RequestTarget: req1[1],
		HttpVersion:   versionPart[1],
	}
	return r, restOfTheString, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	n, err := io.ReadAll(reader)
	if err != nil {
		return nil, ERROR_BAD_START_LINE

	}

	r, _, e := parseRequestLine(string(n))
	if e != nil {
		return nil, e
	}
	return &Request{
		RequestLine: *r,
	}, e
}
