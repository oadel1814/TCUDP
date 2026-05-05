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
	state    ConnectionState
}

type ConnectionState int

const (
	CLOSED ConnectionState = iota
	SYN_SENT
	SYN_RECEIVED
	ESTABLISHED
)

func NewConnection(udp *udp.UDPConn, peer *net.UDPAddr) *Connection {
	return &Connection{
		seq:      0,
		ack:      0,
		udp:      udp,
		peerAddr: peer,
		state:    CLOSED,
	}
}

// CLIENT side
func (c *Connection) Connect() error {

	syn := &Packet{
		SEQ:   c.seq,
		Flags: SYN,
	}

	// send SYN
	ComputeChecksumAndSet(syn)
	data, _ := syn.encode()
	c.udp.Send(data, c.peerAddr)
	c.state = SYN_SENT

	// wait for SYNACK
	for {
		raw, addr, err := c.udp.Receive()
		if err != nil {
			return err
		}

		pkt, _ := decode(raw)
		if !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&(SYN|ACK) == (SYN | ACK) {
			c.ack = pkt.SEQ + 1
			c.peerAddr = addr
			break
		}
	}

	// send ACK
	ack_pkt := &Packet{
		SEQ:   c.seq + 1,
		ACK:   c.ack,
		Flags: ACK,
	}
	ComputeChecksumAndSet(ack_pkt)
	data, _ = ack_pkt.encode()
	c.udp.Send(data, c.peerAddr)
	c.state = ESTABLISHED

	return nil
}

// SERVER side
func (c *Connection) Listen() (*Connection, error) {
	// 1. Wait for SYN
	// 2. Send SYNACK
	// 3. Wait for ACK
	// 4. Set state = StateEstablished
}

func (c *Connection) Send(data []byte) error {
	// 1. Build packet
	// 2. ComputeChecksumAndSet
	// 3. Send & wait for ACK (stop-and-wait)
	// 4. Retransmit on timeout
}

func (c *Connection) Receive() ([]byte, error) {
	// 1. Read packet
	// 2. VerifyChecksum
	// 3. Send ACK back
	// 4. Return data
}

func (c *Connection) Close() error {
	// 1. Send FIN
	// 2. Wait for ACK
	// 3. Set state = StateClosed
}
