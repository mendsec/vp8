package vp8

import "testing"

func TestRefFrameManager_copyLastToAltRef(t *testing.T) {
	m := newRefFrameManager(16, 16)

	// Should not panic or do anything if last is invalid
	m.copyLastToAltRef()

	// Make last valid
	m.last = m.allocBuffer()
	m.last.valid = true
	m.last.Y[0] = 42
	m.last.Cb[0] = 43
	m.last.Cr[0] = 44

	// First time
	m.copyLastToAltRef()
	if !m.altRef.valid {
		t.Error("altRef should be valid")
	}
	if m.altRef.Y[0] != 42 || m.altRef.Cb[0] != 43 || m.altRef.Cr[0] != 44 {
		t.Error("Data not copied to altRef")
	}

	// Change last and copy again
	m.last.Y[0] = 99
	m.copyLastToAltRef()
	if m.altRef.Y[0] != 99 {
		t.Error("Data not updated in altRef")
	}
}
