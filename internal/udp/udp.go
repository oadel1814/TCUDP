package udp

import (
	"errors"
	"net"
	"time"
)

type UDPConn struct {
	conn *net.UDPConn
}

func NewUDPConn(addr string) (*UDPConn, error) {

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, errors.New("connection error!")
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, errors.New("connection error!")
	}

	return &UDPConn{conn: conn}, nil
}

// raw udp sending and receiving without reliability checks
func (u *UDPConn) Send(data []byte, addr *net.UDPAddr) error {
	_, err := u.conn.WriteToUDP(data, addr)
	if err != nil {
		return errors.New("send error!")
	}
	return nil
}

func (u *UDPConn) Receive() ([]byte, *net.UDPAddr, error) {
	buf := make([]byte, 65507)
	n, addr, err := u.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, err
	}
	return buf[:n], addr, nil
}

// Set a deadline for timeout support
func (u *UDPConn) SetTimeout(d time.Duration) {
	if d == 0 {
		u.conn.SetReadDeadline(time.Time{})
	} else {
		u.conn.SetReadDeadline(time.Now().Add(d * time.Millisecond))
	}
}

// close the connection
func (u *UDPConn) Close() {
	u.conn.Close()
}
