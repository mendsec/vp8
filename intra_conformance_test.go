package vp8

import (
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Key frames were validated against golang.org/x/image/vp8, which decodes them
// and is therefore a real check -- but it is a decoder that never sees an
// inter frame, and quality was never measured at all. On content with any
// texture the open-loop analysis produced key frames that decoded at 13 dB
// while their neighbours decoded at 34, and nothing in the suite noticed.
//
// vpxdec runs as a test-time subprocess, never linked, so the package stays
// cgo-free. The test skips when it is not installed:
//
//	apt-get install vpx-tools
const (
	intraWidth  = 128
	intraHeight = 96
	intraFPS    = 30
)

// intraSourceFrame is a textured pattern with structured chroma. Flat content
// hides this defect completely: open-loop prediction is exact when there is
// nothing for the quantiser to round away.
func intraSourceFrame(i int) []byte {
	buf := make([]byte, 0, intraWidth*intraHeight*3/2)
	for y := 0; y < intraHeight; y++ {
		for x := 0; x < intraWidth; x++ {
			v := (x/16+y/16)%2*80 + 60
			if bar := (i * 4) % intraWidth; x >= bar && x < bar+16 {
				v = 235
			}
			buf = append(buf, byte(v))
		}
	}
	for _, base := range []int{95, 165} {
		for y := 0; y < intraHeight/2; y++ {
			for x := 0; x < intraWidth/2; x++ {
				buf = append(buf, byte(base+(x/8+y/8+i)%2*35))
			}
		}
	}
	return buf
}

// writeIntraIVF wraps the frames in the container vpxdec reads.
func writeIntraIVF(t *testing.T, frames [][]byte) string {
	t.Helper()
	hdr := make([]byte, 32)
	copy(hdr[0:], "DKIF")
	binary.LittleEndian.PutUint16(hdr[4:], 0)  // version
	binary.LittleEndian.PutUint16(hdr[6:], 32) // header length
	copy(hdr[8:], "VP80")
	binary.LittleEndian.PutUint16(hdr[12:], intraWidth)
	binary.LittleEndian.PutUint16(hdr[14:], intraHeight)
	binary.LittleEndian.PutUint32(hdr[16:], intraFPS)
	binary.LittleEndian.PutUint32(hdr[20:], 1)
	binary.LittleEndian.PutUint32(hdr[24:], uint32(len(frames)))

	out := hdr
	for i, f := range frames {
		fh := make([]byte, 12)
		binary.LittleEndian.PutUint32(fh[0:], uint32(len(f)))
		binary.LittleEndian.PutUint64(fh[4:], uint64(i))
		out = append(out, fh...)
		out = append(out, f...)
	}

	path := filepath.Join(t.TempDir(), "stream.ivf")
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatalf("writing the IVF: %v", err)
	}
	return path
}

func intraLumaPSNR(got, want []byte) float64 {
	var sum float64
	for i := range got {
		d := float64(int(got[i]) - int(want[i]))
		sum += d * d
	}
	mse := sum / float64(len(got))
	if mse == 0 {
		return math.Inf(1)
	}
	return 10 * math.Log10(255*255/mse)
}

// TestKeyFramesReconstructFaithfully encodes a run of key frames on textured
// content and checks that libvpx gets back what was encoded.
//
// Before prediction was closed, the same run measured 12.95 to 34.00 dB,
// swinging by twenty decibels between frames that differ only in where a bar
// sits. The swing is the signature: a frame whose macroblocks happen to
// predict well is fine, and one where the error accumulates is not.
func TestKeyFramesReconstructFaithfully(t *testing.T) {
	requireVpxdec(t)

	enc, err := NewEncoder(intraWidth, intraHeight, intraFPS)
	if err != nil {
		t.Fatalf("NewEncoder: %v", err)
	}
	enc.SetKeyFrameInterval(0) // every frame is a key frame

	const count = 10
	frames := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		frame, err := enc.Encode(intraSourceFrame(i))
		if err != nil {
			t.Fatalf("Encode frame %d: %v", i, err)
		}
		if frame[0]&1 != 0 {
			t.Fatalf("frame %d is not a key frame", i)
		}
		frames = append(frames, frame)
	}

	rawPath := filepath.Join(t.TempDir(), "decoded.i420")
	cmd := exec.Command("vpxdec", "--i420", "-o", rawPath, writeIntraIVF(t, frames))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the reference decoder rejected the encoder's output: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("reading the decoded frames: %v", err)
	}

	frameSize := intraWidth * intraHeight * 3 / 2
	if got := len(raw) / frameSize; got != count {
		t.Fatalf("decoded %d frames, want %d", got, count)
	}

	worst := math.Inf(1)
	for i := 0; i < count; i++ {
		psnr := intraLumaPSNR(raw[i*frameSize:i*frameSize+intraWidth*intraHeight],
			intraSourceFrame(i)[:intraWidth*intraHeight])
		t.Logf("key frame %d: luma PSNR %.2f dB", i, psnr)
		if psnr < worst {
			worst = psnr
		}
		if psnr < 35 {
			t.Errorf("key frame %d: luma PSNR %.2f dB is too low for an intra frame at this quantiser",
				i, psnr)
		}
	}
	t.Logf("worst key frame: %.2f dB", worst)
}
