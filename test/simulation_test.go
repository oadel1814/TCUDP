package test

import (
	"TCUDP/internal/transport"
	"testing"
)

func TestSimulationCorruption(t *testing.T) {
	data := []byte("Hello World!")
	corrupted := transport.SimulateCorruption(data)
	if len(corrupted) == 0 {
		t.Error("Corrupted data is empty")
	}
	if len(corrupted) != len(data) {
		t.Errorf("Expected length %d, got %d", len(data), len(corrupted))
	}
}

func TestSimulateLoss(t *testing.T) {
	// The function relies on rand, so we just make sure it returns a valid bool and doesn't panic
	_ = transport.SimulateLoss()
}
