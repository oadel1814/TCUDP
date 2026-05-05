package transport

import "math/rand"

func SimulateLoss() bool {
	// random loss with 10% probability
	return rand.Float32() < 0.1
}
func SimulateCorruption(data []byte) []byte {
	// random corruption with 10% probability
	if rand.Float32() < 0.1 && len(data) > 0 {
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		corrupted[rand.Intn(len(data))] ^= 0xFF // flip a random byte
		return corrupted
	}
	return data
}
