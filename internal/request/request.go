package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"boot.theprimeagen.tv/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit   parserState = "init"
	StateHeader parserState = "Header"
	StateDone   parserState = "done"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
}

func parseRequestLine(httpRequest []byte) (*RequestLine, int, error) {

	idx := bytes.Index(httpRequest, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}
	startLineByte := httpRequest[:idx]
	read := idx + len(SEPARATOR)
	startString := string(startLineByte)
	req1 := strings.Split(startString, " ")
	if len(req1) != 3 {
		return nil, 0, ERROR_BAD_START_LINE
	}

	versionPart := strings.Split(req1[2], "/")
	if len(versionPart) != 2 || versionPart[0] != "HTTP" {
		return nil, 0, ERROR_BAD_START_LINE
	}

	r := &RequestLine{
		Method:        req1[0],
		RequestTarget: req1[1],
		HttpVersion:   versionPart[1],
	}
	return r, read, nil
}

func parseHeaderLine(headers []byte) (int, error) {
	idx := bytes.Index(headers, SEPARATOR)
	if idx == -1 {
		return 0, nil
	}
	return idx, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			r.state = StateHeader
			read += n
			break
		case StateHeader:

			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			if done {
				r.state = StateDone
			}
			read += n
		case StateDone:
			break outer
		default:
			panic("parsing logic failed")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ERROR_BAD_START_LINE = fmt.Errorf("start line is bad")
var SEPARATOR = []byte("\r\n")

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, ERROR_BAD_START_LINE
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, ERROR_BAD_START_LINE
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil
}
