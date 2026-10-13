<div align="center">

  <h1>vp8</h1>
  <p><strong>A VP8 encoder and decoder written entirely in Go. No cgo, no libvpx, no FFmpeg.</strong></p>

  <!-- Badges -->
  [![Go Reference](https://pkg.go.dev/badge/github.com/mendsec/vp8.svg)](https://pkg.go.dev/github.com/mendsec/vp8)
  [![Lint](https://github.com/mendsec/vp8/actions/workflows/lint.yml/badge.svg)](https://github.com/mendsec/vp8/actions/workflows/lint.yml)
  [![Test](https://github.com/mendsec/vp8/actions/workflows/test.yml/badge.svg)](https://github.com/mendsec/vp8/actions/workflows/test.yml)
  [![CI](https://github.com/mendsec/vp8/actions/workflows/ci.yml/badge.svg)](https://github.com/mendsec/vp8/actions/workflows/ci.yml)
  [![DevSecOps](https://github.com/mendsec/vp8/actions/workflows/devsecops.yml/badge.svg)](https://github.com/mendsec/vp8/actions/workflows/devsecops.yml)
  ![Go Version](https://img.shields.io/badge/Go-1.27.0-blue)
  ![Coverage](https://img.shields.io/badge/Coverage-91.7%25-brightgreen)
  [![License](https://img.shields.io/github/license/mendsec/vp8)](https://github.com/mendsec/vp8/blob/develop/LICENSE)
</div>

<br>

`vp8` covers both ends of a VP8 video stream (RFC 6386) in pure Go: the **`Encoder`** turns raw I420 frames into a bitstream you can hand to WebRTC, and the **`Decoder`** turns any VP8 stream back into `image.YCbCr` pictures, byte for byte the same as libvpx.

It was built for real-time screen and camera streaming (remote desktop, cloud gaming, WebRTC with [Pion](https://github.com/pion/webrtc)) where a C toolchain is a liability: `GOOS=windows go build` just works, the binary runs in a `scratch` container, and memory bugs stay in Go's safety net.

---

## Contents

- [Why pure Go](#why-pure-go)
- [Features](#features)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Correctness](#correctness)
- [Performance](#performance)
- [Limitations](#limitations)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Why pure Go

Binding `libvpx` through cgo works, until you ship it:

- Cross-compiling needs a C cross-toolchain for every target.
- Minimal images (`scratch`, `distroless`, `alpine`) need multi-stage builds and the right shared libraries.
- A buffer overrun in C is a segfault (or worse) that Go's runtime cannot catch.

`golang.org/x/image/vp8` is pure Go, but it decodes key frames only, so it cannot show a live stream, where every frame after the first is an inter frame. This package decodes the whole format and encodes it too.

## Features

### Encoder

- Key frames and inter frames, with integer-pel motion estimation (diamond search with zero-motion early exit) and golden-frame references.
- Constant-bitrate rate control that adjusts the quantizer per frame (`SetBitrate`).
- Wavefront parallel processing: macroblock rows are encoded concurrently while respecting their above/left dependencies.
- Multiple token partitions (1, 2, 4 or 8), quantizer deltas, probability updates, and the simple loop filter.
- SAD kernels in AMD64 assembly (SSE2 `PSADBW`), with a generic Go fallback on every other architecture.
- Pre-allocated frame buffers, so a steady stream does not churn the garbage collector.

### Decoder

- The full VP8 format: key and inter frames, last, golden and altref references, hidden frames, sign bias, SPLITMV, segmentation, multiple token partitions, and both the normal and simple loop filters.
- Output identical to libvpx, checked frame by frame (see [Correctness](#correctness)).
- Robust against corrupt input: truncated or damaged frames return an error or a garbage picture, never a panic.
- Returns a standard `*image.YCbCr`, ready for `image/png`, `image/draw` or a GPU upload.

## Installation

```bash
go get github.com/mendsec/vp8
```

Requires Go 1.27 or newer.

## Quick start

### Encoding

Frames go in as I420 (planar YUV 4:2:0): the Y plane, then U, then V, `width*height*3/2` bytes in all.

```go
package main

import (
	"log"

	"github.com/mendsec/vp8"
)

func main() {
	enc, err := vp8.NewEncoder(1280, 720, 30)
	if err != nil {
		log.Fatal(err)
	}
	enc.SetBitrate(2_500_000)    // bits per second
	enc.SetKeyFrameInterval(300) // a key frame every 10 s, inter frames between

	frame := make([]byte, 1280*720*3/2) // from your camera or screen capture
	for i := 0; i < 90; i++ {
		payload, err := enc.Encode(frame)
		if err != nil {
			log.Fatal(err)
		}
		_ = payload // send it: pion/rtp's VP8Payloader, an IVF writer, ...
	}
}
```

> [!NOTE]
> Without `SetKeyFrameInterval`, every frame is a key frame. That is the safe default for lossy transports, but it costs far more bandwidth than inter frames.

`ForceKeyFrame` makes the next frame a key frame, which is what a WebRTC sender does on a PLI or FIR from the receiver.

### Decoding

```go
package main

import (
	"image/png"
	"log"
	"os"

	"github.com/mendsec/vp8"
)

func main() {
	dec := vp8.NewDecoder()
	for i, frame := range receiveFrames() {
		img, err := dec.Decode(frame)
		if err != nil {
			log.Printf("frame %d: %v", i, err) // e.g. vp8.ErrNotKeyFrame before the first key frame
			continue
		}
		if img == nil {
			continue // a hidden (altref) frame: decoded as a reference, not shown
		}
		f, _ := os.Create("frame.png")
		_ = png.Encode(f, img)
		f.Close()
	}
}

func receiveFrames() [][]byte { return nil } // whole frames, in stream order
```

Give `Decode` whole frames in stream order, one call per frame. The returned image belongs to the decoder and is only valid until the next call, so copy it if you need to keep it. A `Decoder` is not safe for concurrent use; use one per stream.

## Correctness

Both halves are tested against libvpx, the reference implementation:

| Test | What it checks |
|---|---|
| `TestDecoderMatchesLibvpxTestVectors` | Decodes libvpx's 18 official conformance vectors (`vp80-00-comprehensive-001` to `-018`, in `testdata/libvpx`) and checks every shown frame against libvpx's published MD5. |
| `TestDecoderMatchesLibvpx` | Encodes streams with this package's `Encoder` (several sizes, partition counts, golden frames, quantizer deltas, sizes that are not whole macroblocks) and requires the `Decoder` to match `vpxdec` byte for byte. |
| `TestInterFramesDecodeInLibvpx`, `TestKeyFramesReconstructFaithfully` | Encoder output decodes in libvpx to the picture that was encoded. |
| `TestDecoderSurvivesCorruptFrames`, `FuzzDecoder` | Every prefix and a spread of bit flips of a real stream, plus a fuzz target: no input may panic. |

The tests that compare against `vpxdec` skip themselves when it is not on `PATH`, except in CI on Linux, where `vpxdec` is installed and required; the conformance vectors need nothing installed. CI runs the whole suite with the race detector on Linux, macOS and Windows, with Go 1.26 and 1.27.

## Performance

Measured with `go test -bench` on an Intel Core i7-12650H (10 cores, 16 threads), Go 1.27.0, linux/amd64:

| Benchmark | Content | Time per frame | Frames per second |
|---|---|---|---|
| Encode 1280×720 | key frames | 22 ms | ~45 |
| Encode 1920×1080 | key frames | 50 ms | ~20 |
| Decode 1920×1080 | key frame, panning scene | 44 ms | ~23 |
| Decode 1920×1080 | inter frames, panning scene | 45 ms | ~22 |
| Decode 1920×1080 | key frame, desktop content | 28 ms | ~35 |
| Decode 1920×1080 | inter frames, desktop content | 4.7 ms | ~210 |

Screen content is where inter frames shine: most of a desktop does not change between frames, and those macroblocks cost almost nothing to encode or decode. The encoder parallelises across cores; the decoder runs on one goroutine per stream.

To reproduce:

```bash
go test -run '^$' -bench 'BenchmarkEncode|BenchmarkDecode' .
make bench   # the encoder's resolution and bitrate matrix, see BENCHMARK.md
```

## Limitations

The decoder handles the whole format. The encoder emits a useful subset of it:

- Motion estimation is integer-pel only; there is no sub-pixel search.
- No SPLITMV, segmentation, altref frames or temporal scalability.
- The loop filter is the simple one; the normal filter is not implemented in the encoder.

See [GAPS.md](GAPS.md) and [ROADMAP.md](ROADMAP.md) for what is planned, and [WIKI.md](WIKI.md) for how the encoder works inside.

## Development

```bash
make test   # go test -v -race ./...
make lint   # golangci-lint (v2 config in .golangci.yml)
make sec    # gosec
make all    # lint, sec, test, build
```

Every pull request runs tests on three operating systems and two Go versions, `golangci-lint` (pinned), `gosec`, a 70% coverage floor, and Dependabot keeps modules and actions current.

## Contributing

Pull requests are welcome, whether it is a faster kernel, a missing encoder feature or a clearer doc.

- Found a bug, or a stream that decodes differently from libvpx? Open an [issue](https://github.com/mendsec/vp8/issues), with the stream attached if you can.
- Questions, ideas and show-and-tell go in [Discussions](https://github.com/mendsec/vp8/discussions).
- For a pull request, branch from `develop`, target `develop` (see [AGENTS.md](AGENTS.md)), and run `make all` before pushing.
- Report security problems privately, as described in [SECURITY.md](SECURITY.md), not in a public issue.

## License

[MIT](LICENSE). The libvpx test vectors in `testdata/libvpx` are © The WebM Project authors under libvpx's BSD-3-Clause license; they are test data only and nothing from them is compiled into the package.
