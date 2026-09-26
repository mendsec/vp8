<div align="center">
  <!-- Sugestão: Crie um logo simples com o Gopher segurando um rolo de filme ou um ícone de vídeo -->
  <img src="https://via.placeholder.com/800x200.png?text=VP8+Pure+Go+Encoder+Banner" alt="VP8 Pure Go Encoder Banner">

  <h1>VP8 Pure Go Encoder</h1>
  <p><strong>A high-performance, ultra-low latency VP8 video encoder written entirely in Go.</strong></p>

  <!-- Badges -->
  <a href="https://pkg.go.dev/github.com/mendsec/vp8"><img src="https://pkg.go.dev/badge/github.com/mendsec/vp8.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/mendsec/vp8"><img src="https://goreportcard.com/badge/github.com/mendsec/vp8" alt="Go Report Card"></a>
  <a href="https://github.com/mendsec/vp8/actions/workflows/ci.yml"><img src="https://github.com/mendsec/vp8/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/mendsec/vp8/actions/workflows/devsecops.yml"><img src="https://github.com/mendsec/vp8/actions/workflows/devsecops.yml/badge.svg" alt="DevSecOps"></a>
  ![Go Version](https://img.shields.io/badge/Go-1.27.0-blue)
  ![Coverage](https://img.shields.io/badge/Coverage-88.4%25-brightgreen)
  <a href="https://github.com/mendsec/vp8/blob/develop/LICENSE"><img src="https://img.shields.io/github/license/mendsec/vp8" alt="License"></a>
</div>

<br>

Designed specifically for real-time streaming, remote desktop solutions, and cloud gaming applications where installing C compilers or relying on CGO dependencies is prohibitive. By utilizing advanced software engineering techniques natively in Go, it delivers cinematic frame rates while maintaining a completely portable `go build` experience.

---

## 📖 Table of Contents
- [Why Pure Go? (No CGO)](#-why-pure-go-no-cgo)
- [Ecosystem Integrations](#-ecosystem-integrations)
- [Performance & Benchmarks](#-performance--benchmarks)
- [Features & Technical Innovations](#-features--technical-innovations)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [DevSecOps & Quality Assurance](#-devsecops--quality-assurance)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🛑 Why Pure Go? (No CGO)

If you've ever built media applications in Go, you know the pain of using `CGO` to bind libraries like `libvpx`:
- ❌ Cross-compilation becomes a nightmare.
- ❌ Deploying to minimal Docker containers (like `scratch` or `alpine`) requires complex multi-stage builds.
- ❌ Risk of C-level memory leaks and segfaults that Go's garbage collector cannot catch.

**The VP8 Pure Go Encoder solves this.** 
- ✅ `GOOS=windows GOARCH=amd64 go build` just works.
- ✅ Zero external C dependencies.
- ✅ Safe memory architecture by design.

## 🔌 Ecosystem Integrations

Because it is 100% Go, it integrates seamlessly with modern Go networking stacks. It is the perfect pair for **[Pion WebRTC](https://github.com/pion/webrtc)**. Send real-time video over the web without ever touching C bindings or FFmpeg!

## 🚀 Performance & Benchmarks

This encoder uses **Wavefront Parallel Processing (WPP)**, **SIMD Assembly (AVX2/SSE2)**, and **Dynamic CBR Rate Control** to rival the throughput commonly found in industry-standard hardware-accelerated streams.

*Tested on a 16-core Linux AMD64 environment:*
- **480p (854x480)**: > 150 FPS
- **720p (1280x720)**: > 90 FPS
- **1080p (1920x1080)**: > 40 FPS *(Real-time cinematic grade)*

> **Note:** Results scale based on available CPU cores and SIMD instructions (`PSADBW`). For non-AMD64 targets like ARM or WASM, the compiler falls back transparently to highly optimized generic Go routines.

## 🧠 Features & Technical Innovations

For a deep dive into how this encoder breaks the performance barriers of traditional pure-Go media libraries, check out our [Technical Wiki](WIKI.md).

- **Zero-Allocation Pipeline**: Pre-allocated structures (`reconBuf` and `mbsBuf`) completely eliminate garbage collection (GC) spikes frame-by-frame.
- **Wavefront Asynchronous Processing**: Native Go implementation of macroblock row synchronization using `sync.Cond` instead of expensive thread spin-locks.
- **Ultra-Low Latency Heuristics**: Dynamic predictive pruning skips exhaustive Diamond Searches on static screen areas.
- **CBR Rate Controller**: A feedback loop automatically dials the Quantizer Index (QI) dynamically to adapt to network constraints without stuttering.

## 🛠️ Installation

```bash
go get github.com/mendsec/vp8
```

## ⚡ Quick Start

Here is a simple example of how to initialize the encoder and process a raw YUV frame:

```go
package main

import (
	"log"
	"github.com/mendsec/vp8"
)

func main() {
	// 1. Initialize encoder for 1080p at 30 FPS
	encoder, err := vp8.NewEncoder(1920, 1080, 30)
	if err != nil {
		log.Fatalf("Failed to initialize encoder: %v", err)
	}

	// 2. Target 5 Mbps network throughput
	encoder.SetBitrate(5_000_000)

	// 3. Provide raw YUV420 planar bytes (e.g., from camera or screen capture)
	rawYUV := captureScreen()

	// 4. Encode frame to VP8 payload
	vp8Payload, err := encoder.Encode(rawYUV)
	if err != nil {
		log.Fatalf("Encoding error: %v", err)
	}

	// 5. Transmit payload over WebRTC / RTP / UDP
	transmit(vp8Payload)
}

func captureScreen() []byte { return make([]byte, 1920*1080*3/2) }
func transmit(payload []byte) {}
```

## 🛡️ DevSecOps & Quality Assurance

Security and stability are strictly enforced through a robust DevSecOps culture:
- **SAST**: Automated vulnerability scanning powered by `gosec` on every commit.
- **Linters**: Code quality guaranteed through `golangci-lint` in the CI pipeline.
- **Dependabot**: Continuous automated auditing for upstream package vulnerabilities.
- **Zero CGO**: By eliminating C-bindings, we eliminate entire classes of memory unsafety vulnerabilities (Buffer Overflows, Use-After-Free) at the architectural level.

## 🤝 Contributing

Contributions are heavily welcomed! Whether it's optimizing an assembly routine or improving documentation.

1. Please consult our [Security Policy](SECURITY.md) before submitting patches.
2. Read the [AGENTS.md](AGENTS.md) for workflow rules (e.g., Target the `develop` branch).

You can run the full DevSecOps suite locally using our Makefile:

```bash
make all      # Runs linters, security scanners, tests, and builds
make bench    # Evaluates the SIMD and WPP metrics against your local machine
```

## 💬 Community

- Want to discuss the project? Open a [GitHub Discussion](https://github.com/mendsec/vp8/discussions).
- Found a bug? Open an [Issue](https://github.com/mendsec/vp8/issues).

## 📄 License

This project is licensed under the [MIT License](LICENSE).
