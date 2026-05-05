package transport

import "math/rand"

func SimulateLoss() bool {
	// random loss with 1% probability
	return rand.Float32() < 0.01
}
func SimulateCorruption(data []byte) []byte {
	// random corruption with 1% probability
	if rand.Float32() < 0.01 && len(data) > 0 {
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		corrupted[rand.Intn(len(data))] ^= 0xFF // flip a random byte
		return corrupted
	}
	return data
}
