package transport

import (
	"TCUDP/internal/udp"
	"net"
)

type Connection struct {
	seq      uint32
	ack      uint32
	udp      *udp.UDPConn
	peerAddr *net.UDPAddr
}
