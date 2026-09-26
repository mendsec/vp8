# VP8 Pure Go Encoder

![DevSecOps Pipeline](https://github.com/mendsec/vp8/actions/workflows/devsecops.yml/badge.svg)
![Go Version](https://img.shields.io/github/go-mod/go-version/mendsec/vp8)
![License](https://img.shields.io/github/license/mendsec/vp8)

A high-performance, ultra-low latency VP8 video encoder written **entirely in Go**. 

Designed specifically for real-time streaming, remote desktop solutions, and cloud gaming applications where installing C compilers or relying on CGO dependencies (like `libvpx`) is prohibitive. By utilizing advanced software engineering techniques natively in Go, it delivers cinematic frame rates while maintaining a completely portable `go build` experience.

## 🚀 Performance Highlights

This encoder uses **Wavefront Parallel Processing (WPP)**, **SIMD Assembly (AVX2/SSE2)**, and **Dynamic CBR Rate Control** to rival the throughput commonly found in industry-standard hardware-accelerated streams—all in pure software.

*Tested on a 16-core Linux AMD64 environment:*
*   **480p (854x480)**: > 150 FPS
*   **720p (1280x720)**: > 90 FPS
*   **1080p (1920x1080)**: > 40 FPS (Real-time cinematic grade)

*(Note: Results scale based on available CPU cores and SIMD instructions `PSADBW`. For non-AMD64 targets like ARM or WASM, the compiler falls back transparently to optimized generic Go routines.)*

## 🛡️ DevSecOps & Quality Assurance

Security and stability are strictly enforced through a robust DevSecOps culture:
*   **SAST**: Automated vulnerability scanning powered by `gosec` on every commit.
*   **Linters**: Code quality guaranteed through `golangci-lint` in the CI pipeline.
*   **Dependabot**: Continuous automated auditing for upstream package vulnerabilities.
*   **Zero CGO**: By eliminating C-bindings, we eliminate entire classes of memory unsafety vulnerabilities (Buffer Overflows, Use-After-Free) at the architectural level.

## 🧠 Technical Innovations

To read more about how this encoder breaks the performance barriers of traditional pure-Go media libraries, check out our [Technical Wiki](WIKI.md).

*   **Zero-Allocation Pipeline**: Pre-allocated structures (`reconBuf` and `mbsBuf`) completely eliminate garbage collection (GC) frame-by-frame spikes.
*   **Wavefront Assynchronous Processing**: A native Go implementation of macroblock row synchronization using `sync.Cond` instead of expensive thread spin-locks.
*   **Ultra-Low Latency Heuristics**: Dynamic predictive pruning skips exhaustive Diamond Searches on static screen areas, exponentially saving CPU cycles during desktop streaming.
*   **CBR Rate Controller**: A feedback loop automatically dials the Quantizer Index (QI) dynamically to adapt to network constraints without stuttering.

## 🛠️ Usage

### Installation

```bash
go get github.com/mendsec/vp8
```

### Basic Encoding

```go
package main

import (
	"github.com/mendsec/vp8"
	"log"
)

func main() {
	// Initialize encoder for 1080p at 30 FPS
	encoder, err := vp8.NewEncoder(1920, 1080, 30)
	if err != nil {
		log.Fatalf("Failed to initialize encoder: %v", err)
	}

	// Target 5 Mbps network throughput
	encoder.SetBitrate(5_000_000)

	// Provide raw YUV420 planar bytes
	var rawYUV []byte = captureScreen()

	// Encode frame
	vp8Payload, err := encoder.Encode(rawYUV)
	if err != nil {
		log.Fatalf("Encoding error: %v", err)
	}

	// Transmit vp8Payload over WebRTC / RTP / UDP
	transmit(vp8Payload)
}

func captureScreen() []byte { return make([]byte, 1920*1080*3/2) }
func transmit(payload []byte) {}
```

## 🤝 Contributing

Contributions are heavily welcomed! Please consult our [Security Policy](SECURITY.md) before submitting patches. 
You can run the full DevSecOps suite locally using our Makefile:

```bash
make all      # Runs linters, security scanners, tests, and builds
make bench    # Evaluates the SIMD and WPP metrics against your local machine
```

## License
MIT
