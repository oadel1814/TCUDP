package main

import (
	"TCUDP/internal/http"
	"TCUDP/internal/transport"
	"TCUDP/internal/udp"
	"fmt"
	"log"
	"net"
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Error resolving address: %v", err)
	}

	udpConn, err := udp.NewUDPConn("127.0.0.1:0")
	if err != nil {
		log.Fatalf("Error creating UDP conn: %v", err)
	}
	defer udpConn.Close()

	conn := transport.NewConnection(udpConn, serverAddr)
	fmt.Println("Connecting to server...")
	err = conn.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("Connected!")

	client := http.NewClient(conn)

	fmt.Println("\nSending GET /hello")
	resp, err := client.Get("/hello")
	if err != nil {
		log.Fatalf("GET failed: %v", err)
	}
	fmt.Printf("Received Response: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("Body: %s\n", resp.Body)

	fmt.Println("\nSending POST /echo")
	resp, err = client.Post("/echo", "Hello TCUDP!")
	if err != nil {
		log.Fatalf("POST failed: %v", err)
	}
	fmt.Printf("Received Response: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("Body: %s\n", resp.Body)

	fmt.Println("\nClosing connection...")
	conn.Close()
}
