package transport

import (
	"errors"
	"fmt"
	"io"
	"net"

	"TCUDP/internal/udp"
)

type Connection struct {
	seq      uint32
	ack      uint32
	udp      *udp.UDPConn
	peerAddr *net.UDPAddr
	state    ConnectionState
}

const (
	MAX_RETRIES = 5
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

		// Send SYN
		ComputeChecksumAndSet(syn)
		data, _ := syn.encode()
		c.udp.Send(data, c.peerAddr)

		c.state = SYN_SENT

		// Wait for SYN-ACK
		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue // timeout, retry
		}

		pkt, _ := decode(raw)
		if !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&(SYN|ACK) == (SYN|ACK) && pkt.ACK == c.seq+1 {
			c.ack = pkt.SEQ + 1
			c.peerAddr = addr

			// SYN consumed one seq number
			c.seq++

			// Send final ACK
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

		// Wait for SYN
		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue // timeout, retry
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
		c.seq = 0
		c.ack = pkt.SEQ + 1
		c.state = SYN_RECEIVED

		// Prepare SYN-ACK
		synack := &Packet{
			SEQ:   c.seq,
			ACK:   c.ack,
			Flags: SYN | ACK,
		}

		// Send SYN-ACK and wait for final ACK
		for retries := 0; retries < MAX_RETRIES; retries++ {
			ComputeChecksumAndSet(synack)
			data, _ := synack.encode()
			c.udp.Send(data, c.peerAddr)

			c.udp.SetTimeout(TIMEOUT)
			raw2, addr2, err := c.udp.Receive()
			if err != nil {
				continue // timeout, retry SYN-ACK
			}

			pkt2, _ := decode(raw2)
			if !VerifyChecksum(*pkt2) {
				continue
			}

			if pkt2.Flags&ACK != 0 && pkt2.ACK == c.seq+1 {
				c.peerAddr = addr2

				// SYN consumed one seq number
				c.seq++
				c.ack = pkt2.SEQ

				c.state = ESTABLISHED
				return c, nil
			}
		}
	}

	return nil, errors.New("connection timeout: SYN-ACK retries exceeded")
}

func (c *Connection) Send(data []byte) error {
	pkt := &Packet{
		SEQ:  c.seq,
		Data: data,
	}

	ComputeChecksumAndSet(pkt)
	encoded, err := pkt.encode()
	if err != nil {
		return err
	}

	for {
		c.udp.Send(encoded, c.peerAddr)

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue // timeout, retransmit
		}

		if addr != c.peerAddr {
			continue
		}

		ackPkt, err := decode(raw)
		if err != nil {
			continue
		}

		if !VerifyChecksum(*ackPkt) {
			continue
		}

		if ackPkt.ACK == c.seq+1 {
			c.seq++
			return nil
		}

		if ackPkt.ACK > c.seq+1 {
			return fmt.Errorf("sequence desync: expected ACK %d, got %d", c.seq+1, ackPkt.ACK)
		}

		// ackPkt.ACK <= c.seq: old/duplicate ACK, retransmit
	}
}

func (c *Connection) Receive() ([]byte, error) {
	for {
		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			return nil, fmt.Errorf("receive error: %w", err)
		}

		if addr != c.peerAddr {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		// Peer is closing — hand off to passive close handler
		if pkt.Flags&FIN != 0 {
			if err := c.acceptClose(pkt.SEQ); err != nil {
				return nil, fmt.Errorf("close handshake failed: %w", err)
			}
			return nil, io.EOF // signal clean close to the caller
		}

		if pkt.SEQ != c.seq {
			// Duplicate — re-ACK and discard
			c.sendACK(pkt.SEQ + 1)
			continue
		}

		if err := c.sendACK(c.seq + 1); err != nil {
			return nil, fmt.Errorf("failed to send ACK: %w", err)
		}
		c.seq++
		return pkt.Data, nil
	}
}

func (c *Connection) Close() error {
	if c.state == CLOSED {
		return nil
	}

	// send FIN
	fin := &Packet{SEQ: c.seq, Flags: FIN}
	ComputeChecksumAndSet(fin)
	encoded, err := fin.encode()
	if err != nil {
		return fmt.Errorf("failed to encode FIN: %w", err)
	}

	// wait for ACK of FIN
	acked := false
	for i := 0; i < MAX_RETRIES; i++ {
		c.udp.Send(encoded, c.peerAddr)

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue // timeout, retransmit FIN
		}
		if addr != c.peerAddr {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&ACK != 0 && pkt.ACK == c.seq+1 {
			acked = true
			break
		}
	}

	if !acked {
		// Peer is gone
		c.state = CLOSED
		return nil
	}

	c.seq++

	// wait for peer's FIN
	for i := 0; i < MAX_RETRIES; i++ {
		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			break
		}
		if addr != c.peerAddr {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&FIN != 0 {
			// ACK the peer's FIN
			c.sendACK(pkt.SEQ + 1)
			break
		}
	}

	c.state = CLOSED
	return nil
}

func (c *Connection) acceptClose(finSEQ uint32) error {
	// ACK their FIN
	if err := c.sendACK(finSEQ + 1); err != nil {
		return fmt.Errorf("failed to ACK FIN: %w", err)
	}

	// Send our own FIN
	fin := &Packet{SEQ: c.seq, Flags: FIN}
	ComputeChecksumAndSet(fin)
	encoded, err := fin.encode()
	if err != nil {
		return fmt.Errorf("failed to encode FIN: %w", err)
	}

	for i := 0; i < MAX_RETRIES; i++ {
		c.udp.Send(encoded, c.peerAddr)

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue // timeout, retransmit FIN
		}
		if addr != c.peerAddr {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&ACK != 0 && pkt.ACK == c.seq+1 {
			c.seq++
			break // FIN was acknowledged
		}
	}

	c.state = CLOSED
	return nil
}

func (c *Connection) sendACK(ack uint32) error {
	pkt := &Packet{ACK: ack, Flags: ACK}
	ComputeChecksumAndSet(pkt)
	encoded, err := pkt.encode()
	if err != nil {
		return err
	}
	return c.udp.Send(encoded, c.peerAddr)
}
