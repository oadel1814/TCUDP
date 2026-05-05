package transport

func ComputeChecksum(data []byte) uint16
func VerifyChecksum(pkt Packet) bool
