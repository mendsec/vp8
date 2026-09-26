package vp8

import (
	"testing"
)

func TestEncodeInterFrameHeader(t *testing.T) {
	enc := newBoolEncoder()

	deltas := QuantDeltas{}
	loopFilter := loopFilterParams{}
	mbs := make([]macroblock, 1)

	// It just delegates to encodeInterFrameHeaderWithProbs
	encodeInterFrameHeader(enc, 16, 16, 10, deltas, 1, loopFilter, false, mbs)
	out := enc.flush()

	if len(out) == 0 {
		t.Error("encodeInterFrameHeader produced no output")
	}
}

func TestEncodeMVComponent(t *testing.T) {
	enc := newBoolEncoder()
	var probs [19]uint8
	for i := range probs {
		probs[i] = 128
	}

	for v := -15; v <= 15; v++ {
		encodeMVComponent(enc, int16(v), probs)
	}
	out := enc.flush()
	if len(out) == 0 {
		t.Error("encodeMVComponent produced no output")
	}
}
