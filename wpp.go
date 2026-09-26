package vp8

import "sync"

// rowTracker synchronizes Wavefront Parallel Processing (WPP) rows
// without burning CPU cycles in tight spin-loops.
type rowTracker struct {
	mu   sync.Mutex
	cond *sync.Cond
	col  int
}

func newRowTrackers(mbH int) []rowTracker {
	trackers := make([]rowTracker, mbH)
	for i := 0; i < mbH; i++ {
		trackers[i].cond = sync.NewCond(&trackers[i].mu)
		trackers[i].col = -1
	}
	return trackers
}
