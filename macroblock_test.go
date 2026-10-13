package vp8

import "testing"

func TestProcessMacroblockSimple(t *testing.T) {
	mb := processMacroblockSimple()
	if mb.lumaMode != DC_PRED {
		t.Errorf("lumaMode = %v, want DC_PRED", mb.lumaMode)
	}
	if mb.chromaMode != DC_PRED_CHROMA {
		t.Errorf("chromaMode = %v, want DC_PRED_CHROMA", mb.chromaMode)
	}
	if !mb.skip {
		t.Error("skip should be true")
	}
	if mb.dcValue != 0 {
		t.Errorf("dcValue = %d, want 0", mb.dcValue)
	}
}
