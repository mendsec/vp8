package vp8

import "testing"

func TestRateController_Update(t *testing.T) {
	rc := NewRateController(1000000, 30)
	// Initial QI is typically 60

	// Target per frame = 1M / 30 / 8 = 4166 bytes
	// Buffer size = 4166 * 30 = 125000 bytes

	// Force bufferLevel > bufferSize
	rc.Update(200000)
	if rc.bufferLevel != rc.bufferSize {
		t.Errorf("bufferLevel should be capped at bufferSize")
	}

	// 75% threshold
	rc.bufferLevel = int(float64(rc.bufferSize) * 0.8)
	rc.Update(rc.targetPerFrame) // keep buffer steady

	// 60% threshold
	rc.bufferLevel = int(float64(rc.bufferSize) * 0.65)
	rc.Update(rc.targetPerFrame)

	// 40% threshold
	rc.bufferLevel = int(float64(rc.bufferSize) * 0.35)
	rc.Update(rc.targetPerFrame)

	// 25% threshold
	rc.bufferLevel = int(float64(rc.bufferSize) * 0.2)
	rc.Update(rc.targetPerFrame)

	// <= 0
	rc.bufferLevel = -100
	rc.Update(rc.targetPerFrame)
	if rc.bufferLevel != 0 {
		t.Errorf("bufferLevel should be bounded to 0")
	}

	// Test bounds 4 and 127
	rc.currentQI = 0
	rc.Update(rc.targetPerFrame)
	if rc.currentQI != 4 {
		t.Errorf("currentQI should be clamped to 4, got %d", rc.currentQI)
	}

	rc.currentQI = 200
	rc.Update(rc.targetPerFrame)
	if rc.currentQI != 127 {
		t.Errorf("currentQI should be clamped to 127, got %d", rc.currentQI)
	}
}
