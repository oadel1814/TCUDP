package test

import (
	"TCUDP/internal/transport"
	"testing"
)

func TestChecksum(t *testing.T) {
	pkt := transport.Packet{
		SEQ:   1,
		ACK:   2,
		Flags: transport.SYN,
		Data:  []byte("test data for checksum validation"),
	}

	err := transport.ComputeChecksumAndSet(&pkt)
	if err != nil {
		t.Fatalf("Compute checksum failed: %v", err)
	}

	if !transport.VerifyChecksum(pkt) {
		t.Error("Checksum verification failed for valid packet")
	}

	// Tamper data to simulate corruption
	pkt.Data[0] ^= 0xFF
	if transport.VerifyChecksum(pkt) {
		t.Error("Checksum verification should fail for tampered packet")
	}
}
