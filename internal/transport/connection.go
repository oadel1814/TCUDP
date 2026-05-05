package transport

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

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
	TIMEOUT     = 1 * time.Second
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

func sameAddr(a, b *net.UDPAddr) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Port == b.Port && a.IP.Equal(b.IP)
}

func (c *Connection) Connect() error {
	syn := &Packet{
		SEQ:   c.seq,
		Flags: SYN,
	}

	for attempts := 0; attempts < MAX_RETRIES; attempts++ {

		ComputeChecksumAndSet(syn)
		data, _ := syn.encode()
		c.networkSend(data, c.peerAddr)

		c.state = SYN_SENT

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue
		}

		if !sameAddr(addr, c.peerAddr) {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&(SYN|ACK) == (SYN|ACK) && pkt.ACK == c.seq+1 {
			c.ack = pkt.SEQ + 1
			c.seq++ // SYN consumed

			ackPkt := &Packet{
				SEQ:   c.seq,
				ACK:   c.ack,
				Flags: ACK,
			}

			ComputeChecksumAndSet(ackPkt)
			data, _ = ackPkt.encode()
			c.networkSend(data, addr)

			c.peerAddr = addr
			c.state = ESTABLISHED
			return nil
		}
	}

	return errors.New("connection timeout: SYN retries exceeded")
}

func (c *Connection) Listen() (*Connection, error) {
	for {
		c.udp.SetTimeout(TIMEOUT)

		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&SYN == 0 {
			continue
		}

		c.peerAddr = addr
		c.seq = 0
		c.ack = pkt.SEQ + 1
		c.state = SYN_RECEIVED

		synack := &Packet{
			SEQ:   c.seq,
			ACK:   c.ack,
			Flags: SYN | ACK,
		}

		success := false

		for retries := 0; retries < MAX_RETRIES; retries++ {

			ComputeChecksumAndSet(synack)
			data, _ := synack.encode()
			c.networkSend(data, c.peerAddr)

			c.udp.SetTimeout(TIMEOUT)
			raw2, addr2, err := c.udp.Receive()
			if err != nil {
				continue
			}

			if !sameAddr(addr2, c.peerAddr) {
				continue
			}

			pkt2, err := decode(raw2)
			if err != nil || !VerifyChecksum(*pkt2) {
				continue
			}

			if pkt2.Flags&ACK != 0 &&
				pkt2.ACK == c.seq+1 &&
				pkt2.SEQ == c.ack {

				c.seq++
				c.ack = pkt2.SEQ + 1
				c.state = ESTABLISHED
				success = true
				break
			}
		}

		if success {
			return c, nil
		}

		c.state = CLOSED
	}
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
		c.networkSend(encoded, c.peerAddr)

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue
		}

		if !sameAddr(addr, c.peerAddr) {
			continue
		}

		ackPkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*ackPkt) {
			continue
		}

		if ackPkt.ACK == c.seq+1 {
			c.seq++
			return nil
		}

		if ackPkt.ACK > c.seq+1 {
			return fmt.Errorf("sequence desync: expected ACK %d, got %d", c.seq+1, ackPkt.ACK)
		}
	}
}

func (c *Connection) Receive() ([]byte, error) {
	for {
		c.udp.SetTimeout(0) // blocking

		raw, addr, err := c.udp.Receive()
		if err != nil {
			return nil, fmt.Errorf("receive error: %w", err)
		}

		if !sameAddr(addr, c.peerAddr) {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&FIN != 0 {
			if err := c.acceptClose(pkt.SEQ); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}

		// duplicate
		if pkt.SEQ < c.seq {
			c.sendACK(c.seq)
			continue
		}

		// out of order (not supported)
		if pkt.SEQ > c.seq {
			continue
		}

		// correct packet
		if err := c.sendACK(c.seq + 1); err != nil {
			return nil, err
		}

		c.seq++
		return pkt.Data, nil
	}
}

func (c *Connection) Close() error {
	if c.state == CLOSED {
		return nil
	}

	fin := &Packet{SEQ: c.seq, Flags: FIN}
	ComputeChecksumAndSet(fin)
	encoded, err := fin.encode()
	if err != nil {
		return err
	}

	acked := false

	for i := 0; i < MAX_RETRIES; i++ {
		c.networkSend(encoded, c.peerAddr)

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue
		}

		if !sameAddr(addr, c.peerAddr) {
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
		c.state = CLOSED
		return nil
	}

	c.seq++

	// wait for FIN
	for i := 0; i < MAX_RETRIES; i++ {
		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			break
		}

		if !sameAddr(addr, c.peerAddr) {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&FIN != 0 {
			c.sendACK(pkt.SEQ + 1)
			break
		}
	}

	c.state = CLOSED
	return nil
}

func (c *Connection) acceptClose(finSEQ uint32) error {
	if err := c.sendACK(finSEQ + 1); err != nil {
		return err
	}

	fin := &Packet{SEQ: c.seq, Flags: FIN}
	ComputeChecksumAndSet(fin)
	encoded, err := fin.encode()
	if err != nil {
		return err
	}

	for i := 0; i < MAX_RETRIES; i++ {
		c.networkSend(encoded, c.peerAddr)

		c.udp.SetTimeout(TIMEOUT)
		raw, addr, err := c.udp.Receive()
		if err != nil {
			continue
		}

		if !sameAddr(addr, c.peerAddr) {
			continue
		}

		pkt, err := decode(raw)
		if err != nil || !VerifyChecksum(*pkt) {
			continue
		}

		if pkt.Flags&ACK != 0 && pkt.ACK == c.seq+1 {
			c.seq++
			break
		}
	}

	c.state = CLOSED
	return nil
}

func (c *Connection) sendACK(ack uint32) error {
	pkt := &Packet{
		ACK:   ack,
		Flags: ACK,
	}

	ComputeChecksumAndSet(pkt)
	encoded, err := pkt.encode()
	if err != nil {
		return err
	}

	return c.networkSend(encoded, c.peerAddr)
}

// networkSend wraps the UDP send operation with simulated packet loss and corruption
// to ensure the reliable transport logic handles these cases robustly.
func (c *Connection) networkSend(data []byte, addr *net.UDPAddr) error {
	if SimulateLoss() {
		fmt.Printf("[Simulation] Dropping packet to %s\n", addr.String())
		return nil
	}

	corrupted := SimulateCorruption(data)

	// Check if data actually got mutated to log it
	if len(corrupted) > 0 && len(data) > 0 {
		isCorrupted := false
		for i := 0; i < len(data); i++ {
			if corrupted[i] != data[i] {
				isCorrupted = true
				break
			}
		}
		if isCorrupted {
			fmt.Printf("[Simulation] Corrupted packet to %s\n", addr.String())
		}
	}

	return c.udp.Send(corrupted, addr)
}
