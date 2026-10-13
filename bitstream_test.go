package vp8

import (
	"testing"
)

func TestEncodeFrameHeader(t *testing.T) {
	enc := newBoolEncoder()

	deltas := QuantDeltas{}
	loopFilter := loopFilterParams{}
	mbs := []macroblock{{}}

	// encodeFrameHeader calls encodeFrameHeaderWithProbs
	encodeFrameHeader(enc, 16, 16, 10, deltas, 1, loopFilter, 1, mbs)
	out := enc.flush()

	if len(out) == 0 {
		t.Error("encodeFrameHeader produced no output")
	}
}

func TestEncodeBPredModes(t *testing.T) {
	enc := newBoolEncoder()

	var modes [16]intraBMode
	for i := range modes {
		modes[i] = B_DC_PRED
	}

	encodeBPredModes(enc, modes)
	out := enc.flush()

	if len(out) == 0 {
		t.Error("encodeBPredModes produced no output")
	}
}

func TestEncodeResidualPartitions(t *testing.T) {
	mbs := []macroblock{{}} // 1 MB

	parts := encodeResidualPartitions(0, mbs, 1, 1) // 0 means OnePartition
	if len(parts) != 1 {
		t.Errorf("expected 1 partition, got %d", len(parts))
	}
}
