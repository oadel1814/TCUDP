package http

import (
	"TCUDP/internal/transport"
	"fmt"
)

type Client struct {
	Conn *transport.Connection
}

func NewClient(conn *transport.Connection) *Client {
	return &Client{Conn: conn}
}

func (c *Client) Do(req *Request) (*Response, error) {
	data := req.Encode()
	err := c.Conn.Send(data)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	respData, err := c.Conn.Receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return ParseResponse(respData)
}

func (c *Client) Get(path string) (*Response, error) {
	req := &Request{
		Method:  "GET",
		Path:    path,
		Headers: make(map[string]string),
	}
	return c.Do(req)
}

func (c *Client) Post(path string, body string) (*Response, error) {
	req := &Request{
		Method: "POST",
		Path:   path,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: body,
	}
	return c.Do(req)
}
