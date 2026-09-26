# VP8 Encoder Benchmark Suite

This project includes a comprehensive benchmark suite designed to test the pure-Go VP8 encoder's performance across various resolutions, quality settings, and configurations (Keyframe-only vs Inter-frame).

## Running the Benchmark

You can run the benchmark suite using the provided shell script or directly via Go:

```bash
./benchmark.sh
# or
go run benchmark/main.go
```

## Benchmark Matrix

The suite automatically tests the following configurations to validate throughput and memory efficiency:
- **720p_q50**: 1280x720, Quality 50, Target 2500 kbps (I-frames only)
- **720p_q10**: 1280x720, Quality 10, Target 2500 kbps (I-frames only)
- **720p_q90**: 1280x720, Quality 90, Target 2500 kbps (I-frames only)
- **480p_q50**: 854x480, Quality 50, Target 1500 kbps (I-frames only)
- **1080p_q50**: 1920x1080, Quality 50, Target 5000 kbps (I-frames only)
- **720p_inter**: 1280x720, Quality 50, Target 2500 kbps (I and P-frames with Motion Estimation)
- **1080p_inter**: 1920x1080, Quality 50, Target 5000 kbps (I and P-frames with Motion Estimation)

## Recent Real-Time Streaming Optimizations

To achieve real-time streaming performance (e.g., >20 fps at 1080p and >100 fps at 480p in pure Go), we implemented several techniques commonly found in ultra-low latency streaming systems:

1. **Wavefront Parallel Processing (WPP)**: Instead of a sequential macroblock loop, the encoder now uses a lock-free spin-wait row-based worker pool (Goroutines + `atomic.Int32`). This allows multiple rows of the frame to be encoded concurrently while safely respecting Above/Left spatial dependencies.
2. **Zero-Motion Early Termination**: For desktop streaming, static backgrounds are common. We now evaluate a `ZeroMV` predictor first. If the Sum of Absolute Differences (SAD) is extremely low, we skip the expensive Diamond Search completely.
3. **Intra-Prediction Bypassing**: In P-frames, if the inter-prediction cost is already optimal (SAD < threshold), we completely bypass testing 16x16 and 4x4 intra-prediction modes.
4. **Intra Mode Early Exit**: In I-frames, intra prediction loops through up to 10 modes. If an early mode yields a near-perfect match (SAD < 16), we break the loop and skip the remaining extrapolation calculations.
5. **Zero-Allocation Pipeline**: Pre-allocated structures in the `Encoder` state to eliminate frame-by-frame memory GC spikes.
6. **Branchless SAD Computations**: Bitwise arithmetic replaced `if diff < 0` inside all pixel-level SAD functions.

## Performance Highlights (16 cores)

- **480p**: ~101 FPS
- **720p**: ~45 FPS
- **1080p**: ~21 FPS (Close to real-time playable target without CGO)

Results are automatically exported to `vp8_benchmark_results.csv`.
