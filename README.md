# TCUDP - Reliable Transport over UDP

TCUDP is a custom network protocol implemented in Go that provides reliable data transfer (similar to TCP) over an unreliable UDP layer. This project also includes simulated network conditions (like packet loss or delay) and an HTTP implementation running on top of this custom transport layer.

## Project Structure

- `cmd/client/`: Contains the client application entry point.
- `cmd/server/`: Contains the server application entry point.
- `internal/http/`: Custom HTTP client/server implementation running over the TCUDP transport layer.
- `internal/transport/`: Core logic for the reliable transport protocol (Connections, Packet structures, Checksums, and Network Simulation).
- `internal/udp/`: Basic UDP wraps and utilities.
- `internal/utils/`: Configuration and logging utilities.
- `test/`: Integration and unit tests for the transport protocol, simulation, and HTTP layer.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) installed on your machine.

### Running the Project

1. **Start the Server**
   Open a terminal and run the server component:
   ```bash
   go run cmd/server/main.go
   ```

2. **Start the Client**
   Open a second terminal and run the client component to interact with the server:
   ```bash
   go run cmd/client/main.go
   ```

## Testing

To run the provided unit and simulation tests:
```bash
go test -v ./...
```
