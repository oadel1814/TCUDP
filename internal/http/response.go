package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Response struct {
	StatusCode int
	Status     string
	Headers    map[string]string
	Body       string
}

func ParseResponse(data []byte) (*Response, error) {
	reader := bufio.NewReader(bytes.NewReader(data))

	// Read Status-Line
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed status line")
	}

	statusCode, _ := strconv.Atoi(parts[1])

	resp := &Response{
		StatusCode: statusCode,
		Status:     parts[2],
		Headers:    make(map[string]string),
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
			resp.Headers[headerParts[0]] = headerParts[1]
		}
	}

	// Read Body
	bodyBytes, _ := io.ReadAll(reader)
	resp.Body = string(bodyBytes)

	return resp, nil
}

func (r *Response) Encode() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, r.Status))

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
