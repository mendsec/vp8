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

## Recent Optimizations

The latest updates have significantly improved the encoder's performance:
1. **Zero-Allocation Macroblock Pipeline**: Internal `macroblock` and `refFrameBuffer` structs are now pre-allocated inside the `Encoder` state, eliminating ~140MB of GC allocations per 50 frames.
2. **Branchless SAD Computations**: Replaced conditional branching (`if diff < 0`) in all Sum of Absolute Differences (SAD) functions (`computeMCSAD16x16`, `computeSAD16x16`, `computeSAD8x8`, `computeSAD4x4`) with bitwise branchless arithmetic, resulting in a ~10% FPS throughput increase.

## Metrics Captured

- **Throughput (FPS)**
- **Average Encode Time (ms)**
- **Output Bitrate (kbps)**
- **Memory Allocations (Per Frame)**
- **GC Pauses**

Results are printed to the console and automatically exported to `vp8_benchmark_results.csv`.
