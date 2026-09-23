package vp8

import (
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The inter-frame path had no decode validation: the existing tests check the
// frame-type bit and a minimum byte length, which a stream no decoder accepts
// passes just as easily as one it does. This runs the encoder's output through
// libvpx's own decoder and compares the pictures that come back against the
// source.
//
// vpxdec is a test-time subprocess, never linked in, so the package stays
// cgo-free. The test skips when it is not installed:
//
//	apt-get install vpx-tools
const (
	confWidth  = 128
	confHeight = 96
	confFPS    = 30
)

// confSourceFrame is a moving pattern with structured chroma. Flat chroma is a
// trap here: it lets a stream whose chroma is wrong score well anyway.
func confSourceFrame(i int) []byte {
	buf := make([]byte, 0, confWidth*confHeight*3/2)
	for y := 0; y < confHeight; y++ {
		for x := 0; x < confWidth; x++ {
			v := (x/16+y/16)%2*80 + 60
			if bar := (i * 4) % confWidth; x >= bar && x < bar+16 {
				v = 235
			}
			buf = append(buf, byte(v))
		}
	}
	for _, base := range []int{95, 165} {
		for y := 0; y < confHeight/2; y++ {
			for x := 0; x < confWidth/2; x++ {
				buf = append(buf, byte(base+(x/8+y/8+i)%2*35))
			}
		}
	}
	return buf
}

// writeIVF wraps the frames in the container vpxdec reads.
func writeIVF(t *testing.T, frames [][]byte) string {
	t.Helper()
	hdr := make([]byte, 32)
	copy(hdr[0:], "DKIF")
	binary.LittleEndian.PutUint16(hdr[4:], 0)  // version
	binary.LittleEndian.PutUint16(hdr[6:], 32) // header length
	copy(hdr[8:], "VP80")
	binary.LittleEndian.PutUint16(hdr[12:], confWidth)
	binary.LittleEndian.PutUint16(hdr[14:], confHeight)
	binary.LittleEndian.PutUint32(hdr[16:], confFPS)
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

func lumaPSNR(got, want []byte) float64 {
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

// TestInterFramesDecodeInLibvpx encodes a short sequence with inter frames in
// it and checks that libvpx reconstructs each one faithfully.
//
// Before the mode and motion-vector layer was rewritten against RFC 6386
// §16-18, the same sequence decoded at 9-12 dB: parseable, and a different
// picture from the one that was encoded.
func TestInterFramesDecodeInLibvpx(t *testing.T) {
	requireVpxdec(t)

	enc, err := NewEncoder(confWidth, confHeight, confFPS)
	if err != nil {
		t.Fatalf("NewEncoder: %v", err)
	}
	enc.SetKeyFrameInterval(20)

	const count = 10
	frames := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		frame, err := enc.Encode(confSourceFrame(i))
		if err != nil {
			t.Fatalf("Encode frame %d: %v", i, err)
		}
		frames = append(frames, frame)
	}
	if frames[0][0]&1 != 0 {
		t.Fatal("the first frame is not a key frame")
	}
	inter := 0
	for _, f := range frames[1:] {
		if f[0]&1 == 1 {
			inter++
		}
	}
	if inter == 0 {
		t.Fatal("no inter frames were produced; this test would prove nothing")
	}

	rawPath := filepath.Join(t.TempDir(), "decoded.i420")
	cmd := exec.Command("vpxdec", "--i420", "-o", rawPath, writeIVF(t, frames))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the reference decoder rejected the encoder's output: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("reading the decoded frames: %v", err)
	}

	frameSize := confWidth * confHeight * 3 / 2
	if got := len(raw) / frameSize; got != count {
		t.Fatalf("decoded %d frames, want %d", got, count)
	}

	for i := 0; i < count; i++ {
		kind := "inter"
		if frames[i][0]&1 == 0 {
			kind = "key"
		}
		psnr := lumaPSNR(raw[i*frameSize:i*frameSize+confWidth*confHeight],
			confSourceFrame(i)[:confWidth*confHeight])
		t.Logf("frame %d (%s): luma PSNR %.2f dB", i, kind, psnr)
		if psnr < 25 {
			t.Errorf("frame %d (%s): luma PSNR %.2f dB is too low to be a faithful encode",
				i, kind, psnr)
		}
	}
}

// requireVpxdec locates libvpx's reference decoder, the conformance oracle.
//
// Absent, the test skips -- vpxdec is not something every contributor has
// installed. On CI that is not good enough: the workflow installs vpx-tools
// and an install that quietly breaks would leave the conformance tests
// skipping while the run stays green, which is the one outcome worse than a
// red one. Setting VP8_REQUIRE_VPXDEC there turns absence into a failure.
func requireVpxdec(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("vpxdec"); err != nil {
		if os.Getenv("VP8_REQUIRE_VPXDEC") != "" {
			t.Fatalf("VP8_REQUIRE_VPXDEC is set, but vpxdec is not installed: %v", err)
		}
		t.Skipf("vpxdec is not installed (apt-get install vpx-tools): %v", err)
	}
}
