package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func ParseRequest(data []byte) (*Request, error) {
	reader := bufio.NewReader(bytes.NewReader(data))

	// Read Request-Line
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed request line")
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Headers: make(map[string]string),
	}

	// Read Headers
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			req.Headers[headerParts[0]] = headerParts[1]
		}
	}

	// Read Body
	bodyBytes, _ := io.ReadAll(reader)
	req.Body = string(bodyBytes)

	return req, nil
}

func (r *Request) Encode() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", r.Method, r.Path))

	if r.Body != "" && r.Headers["Content-Length"] == "" {
		r.Headers["Content-Length"] = fmt.Sprintf("%d", len(r.Body))
	}

	for k, v := range r.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buf.WriteString("\r\n")
	buf.WriteString(r.Body)

	return buf.Bytes()
}
