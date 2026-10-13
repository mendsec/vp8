package vp8

import "testing"

func TestEncodeInterYMode(t *testing.T) {
	enc := newBoolEncoder()
	modes := []intraMode{DC_PRED, V_PRED, H_PRED, TM_PRED, B_PRED}
	for _, mode := range modes {
		encodeInterYMode(enc, mode)
	}
	out := enc.flush()
	if len(out) == 0 {
		t.Error("encodeInterYMode produced no output")
	}
}

func TestEncodeInterMBMode(t *testing.T) {
	enc := newBoolEncoder()
	near := nearMVs{}

	// Test Intra MB
	mbIntra := &macroblock{isInter: false, lumaMode: B_PRED, chromaMode: DC_PRED_CHROMA}
	for i := range mbIntra.bModes {
		mbIntra.bModes[i] = B_DC_PRED
	}
	encodeInterMBMode(enc, mbIntra, near, 128, 128, 128)

	// Test Inter MB - Last Frame
	mbInterLast := &macroblock{isInter: true, refFrame: refFrameLast, interMode: mvModeNearestMV}
	encodeInterMBMode(enc, mbInterLast, near, 128, 128, 128)

	// Test Inter MB - Golden Frame
	mbInterGolden := &macroblock{isInter: true, refFrame: refFrameGolden, interMode: mvModeNearMV}
	encodeInterMBMode(enc, mbInterGolden, near, 128, 128, 128)

	// Test Inter MB - AltRef Frame
	mbInterAlt := &macroblock{isInter: true, refFrame: refFrameAltRef, interMode: mvModeNewMV}
	encodeInterMBMode(enc, mbInterAlt, near, 128, 128, 128)

	out := enc.flush()
	if len(out) == 0 {
		t.Error("encodeInterMBMode produced no output")
	}
}
