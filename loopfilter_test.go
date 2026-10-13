package vp8

import "testing"

func TestComputeFilterLimit(t *testing.T) {
	// interiorLimit = level >> (sharpness > 0 ? (sharpness > 4 ? 2 : 1) : 0)
	// level=10, sharpness=0 => interiorLimit = 10, limit = 20 + 10 = 30
	if limit := computeFilterLimit(10, 0); limit != 30 {
		t.Errorf("computeFilterLimit(10, 0) = %d, want 30", limit)
	}

	// level=10, sharpness=2 => interiorLimit = 10 >> 1 = 5, limit = 20 + 5 = 25
	if limit := computeFilterLimit(10, 2); limit != 25 {
		t.Errorf("computeFilterLimit(10, 2) = %d, want 25", limit)
	}
}

func TestFilterPlane(t *testing.T) {
	plane := make([]byte, 8*8)
	// Just fill with some values
	for i := range plane {
		plane[i] = byte(i)
	}
	// Verify it doesn't crash
	filterPlane(plane, 8, 8, 10, 4)

	// Check that it's doing something or at least coverage is hit
}
