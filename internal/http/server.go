package http

import (
	"TCUDP/internal/transport"
	"fmt"
	"log"
)

type HandlerFunc func(*Request) *Response

type Server struct {
	Conn    *transport.Connection
	Handler HandlerFunc
}

func NewServer(conn *transport.Connection, handler HandlerFunc) *Server {
	return &Server{
		Conn:    conn,
		Handler: handler,
	}
}

func (s *Server) Start() error {
	for {
		data, err := s.Conn.Receive()
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("Connection closed by peer")
				return nil
			}
			return fmt.Errorf("error receiving data: %w", err)
		}

		req, err := ParseRequest(data)
		if err != nil {
			log.Printf("Failed to parse request: %v\n", err)
			continue
		}

		resp := s.Handler(req)
		
		if resp == nil {
			resp = &Response{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Headers:    make(map[string]string),
			}
		}

		err = s.Conn.Send(resp.Encode())
		if err != nil {
			log.Printf("Failed to send response: %v\n", err)
		}
	}
}

