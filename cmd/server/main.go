package main

import (
	"TCUDP/internal/http"
	"TCUDP/internal/transport"
	"TCUDP/internal/udp"
	"fmt"
	"log"
)

func main() {
	udpConn, err := udp.NewUDPConn("127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Error creating UDP conn: %v", err)
	}
	defer udpConn.Close()

	fmt.Println("Server listening on 127.0.0.1:8080")

	// Wait for a connection
	listener := transport.NewConnection(udpConn, nil)
	conn, err := listener.Listen()
	if err != nil {
		log.Fatalf("Failed to accept connection: %v", err)
	}
	fmt.Println("Client connected!")

	handler := func(req *http.Request) *http.Response {
		fmt.Printf("Received %s %s\n", req.Method, req.Path)
		
		resp := &http.Response{
			Headers: make(map[string]string),
		}

		if req.Method == "GET" && req.Path == "/hello" {
			resp.StatusCode = 200
			resp.Status = "OK"
			resp.Body = "Hello from TCUDP Server!"
		} else if req.Method == "POST" && req.Path == "/echo" {
			resp.StatusCode = 200
			resp.Status = "OK"
			resp.Body = "Echo: " + req.Body
		} else {
			resp.StatusCode = 404
			resp.Status = "Not Found"
			resp.Body = "Path not found"
		}
		
		resp.Headers["Content-Type"] = "text/plain"
		return resp
	}

	server := http.NewServer(conn, handler)
	err = server.Start()
	if err != nil {
		log.Printf("Server error: %v", err)
	}
	
	fmt.Println("Connection closed")
}
