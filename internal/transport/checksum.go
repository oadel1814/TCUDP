package transport

import (
	"errors"
)

func ComputeChecksum(data []byte) uint16 {
	var sum uint32

	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	for (sum >> 16) > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	return ^uint16(sum)
}

func ComputeChecksumAndSet(p *Packet) error {
	p.Checksum = 0
	encoded, err := p.encode()
	if err != nil {
		return errors.New("checksum error")
	}
	p.Checksum = ComputeChecksum(encoded)
	return nil
}

func VerifyChecksum(pkt Packet) bool {
	checksum := pkt.Checksum
	pkt.Checksum = 0

	encoded, err := pkt.encode()
	if err != nil {
		return false
	}

	newChecksum := ComputeChecksum(encoded)

	if newChecksum != checksum {
		return false
	}

	return true
}
