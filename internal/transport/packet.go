package transport

type Packet struct {
	SEQ      uint32
	ACK      uint32
	Flags    uint8
	Checksum uint16
	Payload  []byte
}
