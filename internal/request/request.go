package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
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
	StateBody   parserState = "Body"
	StateDone   parserState = "done"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
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

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
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
		case StateHeader:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			if done {
				r.state = StateBody
			}
			read += n
		case StateBody:
			if r.Headers.HasKey("Content-Length") {
				length, err := strconv.Atoi(r.Headers.Get("Content-Length"))
				if err != nil {
					return 0, err
				}

				remaining := min(length-len(r.Body), len(currentData))
				r.Body = append(r.Body, currentData[:remaining]...)
				read += remaining
				if len(r.Body) >= length {
					r.state = StateDone
				}

			} else {
				r.state = StateDone
			}
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
		Body:    []byte{},
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
