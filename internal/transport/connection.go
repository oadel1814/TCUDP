package transport

import (
	"TCUDP/internal/udp"
	"errors"
	"net"
)

type Connection struct {
	seq      uint32
	ack      uint32
	udp      *udp.UDPConn
	peerAddr *net.UDPAddr
	state    ConnectionState
}

const (
	MAX_RETRIES = 20
	TIMEOUT     = 1000
)

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

	for attempts := 0; attempts < MAX_RETRIES; attempts++ {

		// send SYN
		ComputeChecksumAndSet(syn)
		data, _ := syn.encode()
		c.udp.Send(data, c.peerAddr)

		c.state = SYN_SENT

		// wait for SYN-ACK
		c.udp.SetTimeout(TIMEOUT)

		raw, addr, err := c.udp.Receive()
		if err != nil {
			// timeout and retry
			continue
		}

		pkt, _ := decode(raw)

		if !VerifyChecksum(*pkt) {
			continue
		}

		// validate SYNACK
		if pkt.Flags&(SYN|ACK) == (SYN|ACK) &&
			pkt.ACK == c.seq+1 {

			c.ack = pkt.SEQ + 1
			c.peerAddr = addr

			// SYN consumed one seq number
			c.seq++

			// 4. Send final ACK
			ackPkt := &Packet{
				SEQ:   c.seq,
				ACK:   c.ack,
				Flags: ACK,
			}

			ComputeChecksumAndSet(ackPkt)
			data, _ = ackPkt.encode()
			c.udp.Send(data, c.peerAddr)

			c.state = ESTABLISHED
			return nil
		}
	}

	return errors.New("connection timeout: SYN retries exceeded")
}

// SERVER side
func (c *Connection) Listen() (*Connection, error) {
	for attempts := 0; attempts < MAX_RETRIES; attempts++ {

		// wait for SYN
		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue // timeout and retry
		}

		pkt, _ := decode(raw)
		if !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&SYN == 0 {
			continue
		}

		// Initialize connection from SYN
		c.peerAddr = addr
		c.seq = 0           // server initial seq number (can be random)
		c.ack = pkt.SEQ + 1 // ACK client’s SYN
		c.state = SYN_RECEIVED

		// Prepare SYN-ACK
		synack := &Packet{
			SEQ:   c.seq,
			ACK:   c.ack,
			Flags: SYN | ACK,
		}

		// send SYN-ACK and wait for the final ACK
		for retries := 0; retries < MAX_RETRIES; retries++ {

			ComputeChecksumAndSet(synack)
			data, _ := synack.encode()
			c.udp.Send(data, c.peerAddr)

			c.udp.SetTimeout(TIMEOUT)
			raw2, addr2, err := c.udp.Receive()
			if err != nil {
				// timeout and retry SYN-ACK
				continue
			}

			pkt2, _ := decode(raw2)
			if !VerifyChecksum(*pkt2) {
				continue
			}

			// Validate final ACK
			if pkt2.Flags&ACK != 0 && pkt2.ACK == c.seq+1 {
				c.peerAddr = addr2

				// SYN consumed one seq number on both sides
				c.seq++
				c.ack = pkt2.SEQ

				c.state = ESTABLISHED
				return c, nil
			}

			// Ignore anything else and keep waiting/resending
		}

		// Failed to complete handshake for this SYN → go back to waiting for a new SYN
	}

	return nil, errors.New("connection timeout: SYN-ACK retries exceeded")
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
