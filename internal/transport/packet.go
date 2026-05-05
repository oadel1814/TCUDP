package transport

import (
	"encoding/binary"
	"errors"
)

type Packet struct {
	SEQ      uint32
	ACK      uint32
	Flags    uint8
	Checksum uint16
	Length   uint16
	Data     []byte

	// header size = 13 bytes
}

func (p *Packet) encode() ([]byte, error) {
	headerSize := 4 + 4 + 1 + 2 + 2
	buf := make([]byte, headerSize+len(p.Data))

	binary.BigEndian.PutUint32(buf[0:], p.SEQ)
	binary.BigEndian.PutUint32(buf[4:], p.ACK)
	buf[8] = p.Flags
	binary.BigEndian.PutUint16(buf[9:], p.Checksum)
	binary.BigEndian.PutUint16(buf[11:], p.Length)

	copy(buf[13:], p.Data)

	return buf, nil
}

func decode(buf []byte) (*Packet, error) {
	if len(buf) < 13 {
		return nil, errors.New("packet is too short")
	}

	var decoded_packet Packet
	decoded_packet.SEQ = binary.BigEndian.Uint32(buf[0:])
	decoded_packet.ACK = binary.BigEndian.Uint32(buf[4:])
	decoded_packet.Flags = buf[8]
	decoded_packet.Checksum = binary.BigEndian.Uint16(buf[9:])
	decoded_packet.Length = binary.BigEndian.Uint16(buf[11:])
	decoded_packet.Data = make([]byte, len(buf[13:]))
	copy(decoded_packet.Data, buf[13:])

	return &decoded_packet, nil
}
