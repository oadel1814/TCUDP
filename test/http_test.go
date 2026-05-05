package test

import (
	"TCUDP/internal/http"
	"testing"
)

func TestParseRequest(t *testing.T) {
	raw := []byte("GET /hello HTTP/1.1\r\nHost: localhost\r\n\r\n")
	req, err := http.ParseRequest(raw)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}
	if req.Method != "GET" || req.Path != "/hello" {
		t.Errorf("Unexpected request parsing result: %+v", req)
	}
}

func TestParseResponse(t *testing.T) {
	raw := []byte("HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello")
	resp, err := http.ParseResponse(raw)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.StatusCode != 200 || resp.Body != "Hello" {
		t.Errorf("Unexpected response parsing result: %+v", resp)
	}
}

func TestEncodeRequest(t *testing.T) {
	req := &http.Request{
		Method: "POST",
		Path:   "/test",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"key":"value"}`,
	}
	
	encoded := string(req.Encode())
	if encoded == "" {
		t.Error("Encoded request is empty")
	}
}
