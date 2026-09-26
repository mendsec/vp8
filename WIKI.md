# VP8 Encoder - Technical Wiki

Welcome to the internal technical wiki for the Pure Go VP8 Encoder. This document outlines the architecture, the memory model, and the advanced performance optimizations that allow the encoder to achieve ultra-low latency real-time streaming entirely in Go.

## 1. Architecture Overview

The encoder is built around a pure Go implementation of the VP8 video coding format (RFC 6386). It is designed specifically for low-latency streaming scenarios (e.g., remote desktop, cloud gaming) where hardware encoding might be unavailable, avoiding the complex dependencies of CGO/`libvpx`.

Key architectural pillars:
- **Zero-Allocation Pipeline**: The `Encoder` struct pre-allocates reconstruction buffers (`reconBuf`) and macroblock state arrays (`mbsBuf`) upon initialization. This prevents garbage collection (GC) pressure during the `Encode` loop.
- **Wavefront Parallel Processing (WPP)**: Instead of locking the entire frame, rows of macroblocks are processed concurrently using lightweight goroutines and lock-free atomic spin-waits.
- **Dynamic Constant Bitrate (CBR)**: A built-in Rate Controller tracks network frame byte consumption and dynamically adjusts the Quantizer Index (QI) to ensure stable throughput.
- **SIMD Assembly**: Critical bottleneck mathematics (like the Sum of Absolute Differences) use Go's native Plan9 Assembly to tap directly into AVX2/SSE2 instructions.

## 2. Advanced Optimizations

To achieve real-time 1080p performance (>20 fps) and 480p performance (>100 fps) on modern processors, we rely on a combination of aggressive software heuristics and hardware intrinsics.

### 2.1 Wavefront Parallel Processing (WPP)
Standard Go channels introduce context-switching latency that is too slow for processing 16x16 macroblocks (which take ~10µs each). Instead, we map one goroutine per frame row. To respect VP8's spatial dependencies (each block depends on the reconstructed pixels of the block "Above" and "Left"), we use an `atomic.Int32` progress tracker. 
Row `Y` spin-waits (`runtime.Gosched()`) until Row `Y-1` finishes processing Column `X+1`, guaranteeing context availability without mutexes.

### 2.2 Rate Distortion Optimization (RDO) Heuristics
True RDO evaluates the exact bit-cost of every possible mode, which destroys FPS. We use "Fast RDO" heuristics:
- **Zero-Motion Early Termination**: If a macroblock's initial (0,0) motion vector yields an exceptionally low SAD against the reference frame, we skip the expensive Diamond Search completely. This is highly effective for desktop streaming where large screen areas are static.
- **Intra-Prediction Bypassing**: In P-frames, if the inter-frame block difference is below a specific threshold, we entirely bypass the evaluation of the 10 intra-prediction modes.

### 2.3 AMD64 Assembly / SIMD (PSADBW)
The `computeSAD16x16`, `8x8`, and `4x4` functions are implemented in `sad_amd64.s`. They utilize the `PSADBW` instruction, which subtracts, absolute-values, and sums 16 bytes of data in a single clock cycle. This turns a 256-iteration scalar loop into 16 vectorized instructions.
**Portability**: The project remains "Pure Go". For architectures other than AMD64 (like ARM64 or WASM), build tags seamlessly fall back to the generic Go implementations in `sad_generic.go`.

## 3. Dynamic Rate Controller (CBR)
The `RateController` intercepts the byte size of each encoded bitstream. It maintains a virtual "bucket" based on the `targetBitrate` and `fps`. 
- If the bucket overflows (high motion/complexity), the controller bumps the Quantizer Index (QI), lowering the visual quality to preserve network stability.
- If the bucket empties (static screen), it lowers the QI, maximizing visual clarity.
This ensures the output bitrate perfectly matches the network constraints over time.

## 4. Future Investigation & Roadmap

The current throughput is spectacular for pure software, but there are always paths to push it further:

1. **Entropy Coder Parallelization**: Currently, writing the residual tokens into the boolean entropy coder is sequential. VP8 supports Token Partitions (`SetPartitionCount`), which allows splitting the residual stream into up to 8 independent boolean writers.
2. **SIMD for DCT / iDCT**: The Walsh-Hadamard Transform (WHT) and `ForwardDCT4x4` are still implemented in pure Go. Vectorizing these transforms using AVX2 will yield another massive CPU reduction.
3. **SIMD for Sub-pixel Filtering**: Motion compensation uses a 6-tap interpolation filter for quarter-pixel accuracy. Writing a `sixTapFilter_amd64.s` using `PMADDWD` would dramatically accelerate P-frame encoding.
4. **ARM64 NEON Assembly**: With Apple Silicon and AWS Graviton becoming standard, porting `sad_amd64.s` to `sad_arm64.s` using NEON instructions (`UABD` + `UADALP`) is a high priority.
