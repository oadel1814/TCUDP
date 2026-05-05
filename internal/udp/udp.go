package udp

import "net"

type UDPConn struct {
	conn *net.UDPConn
}

func (u *UDPConn) Send(data []byte, addr *net.UDPAddr) error
func (u *UDPConn) Receive() ([]byte, *net.Addr, error)
