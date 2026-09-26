package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/opd-ai/vp8"
)

// BenchmarkConfig represents a test configuration
type BenchmarkConfig struct {
	Name           string
	Width          int
	Height         int
	Quality        int // 0-100 or quantizer 0-127
	KeyframeOnly   bool
	EnableMotion   bool
	Threads        int
	TargetBitrate  int // kbps
	NumFrames      int
	RandSeed       int64
}

// BenchmarkResult stores test results
type BenchmarkResult struct {
	Config             BenchmarkConfig
	ThroughputFPS      float64
	AvgEncodeTimeMs    float64
	TotalTime          time.Duration
	OutputBytes        int64
	BitrateKbps        float64
	AllocsPerFrame     int64
	AllocBytesPerFrame int64
	GCPauses           int
	GCTotalTime        time.Duration
	PeakMemoryMB       float64
	Success            bool
	Error              string
}

// Default test matrix for VP8 encoder
var testMatrix = []BenchmarkConfig{
	// Baseline: 720p, quality 50
	{
		Name:          "720p_q50",
		Width:         1280,
		Height:        720,
		Quality:       50,
		KeyframeOnly:  true,
		EnableMotion:  false,
		Threads:       1,
		TargetBitrate: 2500,
		NumFrames:     100,
		RandSeed:      42,
	},
	// Quality variations
	{
		Name:          "720p_q10",
		Width:         1280,
		Height:        720,
		Quality:       10,
		KeyframeOnly:  true,
		EnableMotion:  false,
		Threads:       1,
		TargetBitrate: 2500,
		NumFrames:     100,
		RandSeed:      42,
	},
	{
		Name:          "720p_q90",
		Width:         1280,
		Height:        720,
		Quality:       90,
		KeyframeOnly:  true,
		EnableMotion:  false,
		Threads:       1,
		TargetBitrate: 2500,
		NumFrames:     100,
		RandSeed:      42,
	},
	// Resolution variations
	{
		Name:          "480p_q50",
		Width:         854,
		Height:        480,
		Quality:       50,
		KeyframeOnly:  true,
		EnableMotion:  false,
		Threads:       1,
		TargetBitrate: 1500,
		NumFrames:     100,
		RandSeed:      42,
	},
	{
		Name:          "1080p_q50",
		Width:         1920,
		Height:        1080,
		Quality:       50,
		KeyframeOnly:  true,
		EnableMotion:  false,
		Threads:       1,
		TargetBitrate: 5000,
		NumFrames:     50,
		RandSeed:      42,
	},
	// Inter-frame validation (Motion)
	{
		Name:          "720p_inter",
		Width:         1280,
		Height:        720,
		Quality:       50,
		KeyframeOnly:  false,
		EnableMotion:  true,
		Threads:       1,
		TargetBitrate: 2500,
		NumFrames:     100,
		RandSeed:      42,
	},
}

func main() {
	fmt.Println("🔧 VP8 Pure Go Encoder Benchmark Tool")
	fmt.Println("=====================================")
	fmt.Printf("💻 Go version: %s\n", runtime.Version())
	fmt.Printf("💻 CPU cores: %d\n", runtime.NumCPU())
	fmt.Printf("💻 OS/Arch: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	var results []BenchmarkResult
	totalTests := len(testMatrix)

	for i, config := range testMatrix {
		fmt.Printf("[%d/%d] Testing: %s\n", i+1, totalTests, config.Name)
		printConfig(config)

		result := runBenchmark(config)
		results = append(results, result)

		if result.Success {
			fmt.Printf("   ✅ Throughput: %.2f fps | Avg: %.2f ms/frame | Output: %.2f KB | Bitrate: %.2f kbps\n",
				result.ThroughputFPS, result.AvgEncodeTimeMs, float64(result.OutputBytes)/1024, result.BitrateKbps)
			fmt.Printf("   📊 Allocations: %d allocs/frame, %d bytes/frame | GC: %d pauses, %v total\n",
				result.AllocsPerFrame, result.AllocBytesPerFrame, result.GCPauses, result.GCTotalTime)
		} else {
			fmt.Printf("   ❌ Failed: %s\n", result.Error)
		}
		fmt.Println()
	}

	generateReport(results)
	printSummary(results)
}

func printConfig(config BenchmarkConfig) {
	fmt.Printf("   Resolution: %dx%d | Quality: %d | Threads: %d | Frames: %d\n",
		config.Width, config.Height, config.Quality, config.Threads, config.NumFrames)
	fmt.Printf("   Keyframes only: %v | Motion estimation: %v | Target: %d kbps\n",
		config.KeyframeOnly, config.EnableMotion, config.TargetBitrate)
}

func runBenchmark(config BenchmarkConfig) BenchmarkResult {
	result := BenchmarkResult{
		Config: config,
	}

	runtime.GC()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	gcStart := time.Now()
	oldGCPercent := debugSetGCPercent(-1)
	defer debugSetGCPercent(oldGCPercent)

	encoder, err := createEncoder(config)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to create encoder: %v", err)
		return result
	}

	fmt.Printf("   ⏳ Generating test frames...\n")
	frames := generateTestFrames(config)

	startTime := time.Now()
	var totalBytes int64
	var mu sync.Mutex

	for frameIdx, frame := range frames {
		encoded, err := encoder.Encode(frame)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Encode failed at frame %d: %v", frameIdx, err)
			return result
		}

		mu.Lock()
		totalBytes += int64(len(encoded))
		mu.Unlock()
	}

	totalTime := time.Since(startTime)
	result.TotalTime = totalTime

	runtime.ReadMemStats(&memStatsAfter)
	gcTotalTime := time.Since(gcStart)

	durationSeconds := totalTime.Seconds()
	result.ThroughputFPS = float64(config.NumFrames) / durationSeconds
	result.AvgEncodeTimeMs = (durationSeconds / float64(config.NumFrames)) * 1000
	result.OutputBytes = totalBytes
	result.BitrateKbps = (float64(totalBytes) * 8) / durationSeconds / 1000

	result.AllocsPerFrame = int64(memStatsAfter.Mallocs-memStatsBefore.Mallocs) / int64(config.NumFrames)
	result.AllocBytesPerFrame = int64(memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc) / int64(config.NumFrames)
	result.GCPauses = int(memStatsAfter.NumGC - memStatsBefore.NumGC)
	result.GCTotalTime = gcTotalTime
	result.PeakMemoryMB = float64(memStatsAfter.HeapInuse) / 1024 / 1024

	result.Success = true
	return result
}

func createEncoder(config BenchmarkConfig) (*vp8.Encoder, error) {
	enc, err := vp8.NewEncoder(config.Width, config.Height, 30)
	if err != nil {
		return nil, err
	}
	enc.SetBitrate(config.TargetBitrate * 1000)

	if !config.KeyframeOnly {
		enc.SetKeyFrameInterval(30)
	}

	return enc, nil
}

func generateTestFrames(config BenchmarkConfig) [][]byte {
	frames := make([][]byte, config.NumFrames)
	rng := rand.New(rand.NewSource(config.RandSeed))

	for i := 0; i < config.NumFrames; i++ {
		img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
		offset := i * 10
		drawPattern(img, config.Width, config.Height, offset, rng)
		frames[i] = rgbaToYuv420(img)
	}

	return frames
}

func drawPattern(img *image.RGBA, width, height, offset int, rng *rand.Rand) {
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{20, 20, 30, 255}}, image.Point{}, draw.Src)

	for i := 0; i < 20; i++ {
		x := (int(rng.Int31n(int32(width))) + offset) % width
		y := int(rng.Int31n(int32(height)))
		w := int(rng.Int31n(100)) + 50
		h := int(rng.Int31n(100)) + 50

		r := uint8(rng.Int31n(256))
		g := uint8(rng.Int31n(256))
		b := uint8(rng.Int31n(256))

		rect := image.Rect(x, y, x+w, y+h)
		draw.Draw(img, rect, &image.Uniform{color.RGBA{r, g, b, 255}}, image.Point{}, draw.Src)
	}

	for y := 0; y < height; y += 20 {
		for x := offset % 20; x < width; x += 40 {
			draw.Draw(img, image.Rect(x, y, x+20, y+10), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
		}
	}
}

func clampFloat64(val, minVal, maxVal float64) float64 {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

func rgbaToYuv420(img *image.RGBA) []byte {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	ySize := w * h
	uvSize := (w / 2) * (h / 2)
	yuv := make([]byte, ySize+2*uvSize)

	yPlane := yuv[:ySize]
	uPlane := yuv[ySize : ySize+uvSize]
	vPlane := yuv[ySize+uvSize:]

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.RGBAAt(x, y)
			r, g, b := float64(c.R), float64(c.G), float64(c.B)

			yVal := uint8(clampFloat64(0.299*r+0.587*g+0.114*b, 0, 255))
			yPlane[y*w+x] = yVal

			if y%2 == 0 && x%2 == 0 {
				uVal := uint8(clampFloat64(-0.1687*r-0.3313*g+0.5*b+128, 0, 255))
				vVal := uint8(clampFloat64(0.5*r-0.4187*g-0.0813*b+128, 0, 255))
				uvIdx := (y/2)*(w/2) + (x / 2)
				uPlane[uvIdx] = uVal
				vPlane[uvIdx] = vVal
			}
		}
	}
	return yuv
}

func generateReport(results []BenchmarkResult) {
	csvFile := "vp8_benchmark_results.csv"
	f, err := os.Create(csvFile)
	if err != nil {
		fmt.Printf("⚠️  Error creating CSV: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintln(f, "Config,Resolution,Quality,Threads,KeyframesOnly,Throughput_FPS,AvgTime_Ms,Output_KB,Bitrate_Kbps,Allocs_PerFrame,AllocBytes_PerFrame,GC_Pauses,PeakMemory_MB,Success")

	for _, r := range results {
		status := "OK"
		if !r.Success {
			status = "FAIL"
		}
		fmt.Fprintf(f, "%s,%dx%d,%d,%d,%v,%.2f,%.2f,%.2f,%.2f,%d,%d,%d,%.2f,%s\n",
			r.Config.Name,
			r.Config.Width,
			r.Config.Height,
			r.Config.Quality,
			r.Config.Threads,
			r.Config.KeyframeOnly,
			r.ThroughputFPS,
			r.AvgEncodeTimeMs,
			float64(r.OutputBytes)/1024,
			r.BitrateKbps,
			r.AllocsPerFrame,
			r.AllocBytesPerFrame,
			r.GCPauses,
			r.PeakMemoryMB,
			status,
		)
	}

	fmt.Printf("\n💾 CSV report saved to: %s\n", csvFile)
}

func printSummary(results []BenchmarkResult) {
	fmt.Println("📊 Results Summary")
	fmt.Println("=================")

	var successful []BenchmarkResult
	for _, r := range results {
		if r.Success {
			successful = append(successful, r)
		}
	}

	if len(successful) == 0 {
		fmt.Println("❌ No tests completed successfully")
		return
	}

	var bestThroughput BenchmarkResult
	for _, r := range successful {
		if r.ThroughputFPS > bestThroughput.ThroughputFPS {
			bestThroughput = r
		}
	}

	var bestCompression BenchmarkResult
	bestCompression.BitrateKbps = 999999
	for _, r := range successful {
		if r.BitrateKbps < bestCompression.BitrateKbps && r.Config.Width == 1280 {
			bestCompression = r
		}
	}

	var bestEfficiency BenchmarkResult
	bestScore := 0.0
	for _, r := range successful {
		if r.AllocsPerFrame == 0 {
			r.AllocsPerFrame = 1
		}
		score := r.ThroughputFPS / float64(r.AllocsPerFrame)
		if score > bestScore {
			bestScore = score
			bestEfficiency = r
		}
	}

	fmt.Println("\n🏆 Best Configurations:")
	fmt.Printf("   ⚡ Highest Throughput: %.2f fps (%s, %dx%d)\n",
		bestThroughput.ThroughputFPS, bestThroughput.Config.Name, bestThroughput.Config.Width, bestThroughput.Config.Height)
	fmt.Printf("   📦 Best Compression: %.2f kbps (%s, %dx%d)\n",
		bestCompression.BitrateKbps, bestCompression.Config.Name, bestCompression.Config.Width, bestCompression.Config.Height)
	fmt.Printf("   ⚖️  Best Efficiency: %.6f score (%s, %dx%d)\n",
		bestScore, bestEfficiency.Config.Name, bestEfficiency.Config.Width, bestEfficiency.Config.Height)

	fmt.Println("\n📋 Performance Table:")
	fmt.Printf("%-15s %-10s %-10s %-10s %-12s %-10s %-10s\n",
		"Config", "Throughput", "Avg Time", "Bitrate", "Allocs/Frame", "GC Pauses", "Status")
	fmt.Printf("%-15s %-10s %-10s %-10s %-12s %-10s %-10s\n",
		"---------------", "----------", "----------", "----------", "------------", "----------", "----------")

	for _, r := range successful {
		fmt.Printf("%-15s %-10.1f %-10.1f %-10.1f %-12d %-10d %-10s\n",
			r.Config.Name, r.ThroughputFPS, r.AvgEncodeTimeMs, r.BitrateKbps, r.AllocsPerFrame, r.GCPauses, "✅")
	}
}

func debugSetGCPercent(percent int) int {
	return 100 // Minimal placeholder
}
