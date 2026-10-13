package vp8

// RateController manages Constant Bitrate (CBR) streaming by dynamically
// adjusting the Quantizer Index (QI) based on target buffer levels.
type RateController struct {
	targetBitrate  int
	fps            int
	targetPerFrame int

	bufferSize  int
	bufferLevel int

	currentQI int
}

// NewRateController initializes a CBR controller.
func NewRateController(bitrate, fps int) *RateController {
	if fps <= 0 {
		fps = 30
	}
	targetPerFrame := bitrate / 8 / fps

	// Initial QI based on rough estimate (similar to legacy static mapping)
	ratio := float64(bitrate-100_000) / float64(8_000_000-100_000)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	initialQI := 127 - int(ratio*120) // Use wider range 7..127 for better control

	return &RateController{
		targetBitrate:  bitrate,
		fps:            fps,
		targetPerFrame: targetPerFrame,
		bufferSize:     targetPerFrame * fps,       // 1 second tolerance buffer
		bufferLevel:    (targetPerFrame * fps) / 2, // start at 50% capacity
		currentQI:      initialQI,
	}
}

// Update feeds the size of the last encoded frame to the controller and returns the new QI.
func (rc *RateController) Update(encodedBytes int) int {
	// Delta relative to target
	rc.bufferLevel += encodedBytes - rc.targetPerFrame

	// Adjust QI based on buffer thresholds to prevent network congestion
	if rc.bufferLevel > rc.bufferSize {
		rc.currentQI += 4 // Panic: heavily reduce quality
		rc.bufferLevel = rc.bufferSize
	} else if rc.bufferLevel > int(float64(rc.bufferSize)*0.75) {
		rc.currentQI += 2
	} else if rc.bufferLevel > int(float64(rc.bufferSize)*0.60) {
		rc.currentQI += 1
	} else if rc.bufferLevel <= 0 {
		rc.currentQI -= 4 // Network is free: drastically increase quality
		rc.bufferLevel = 0
	} else if rc.bufferLevel < int(float64(rc.bufferSize)*0.25) {
		rc.currentQI -= 2
	} else if rc.bufferLevel < int(float64(rc.bufferSize)*0.40) {
		rc.currentQI -= 1
	}

	// Clamp QI to valid VP8 bounds (0 is lossless, but typical streaming avoids it)
	if rc.currentQI < 4 {
		rc.currentQI = 4
	}
	if rc.currentQI > 127 {
		rc.currentQI = 127
	}

	return rc.currentQI
}

// GetCurrentQI returns the current suggested Quantizer Index.
func (rc *RateController) GetCurrentQI() int {
	return rc.currentQI
}
