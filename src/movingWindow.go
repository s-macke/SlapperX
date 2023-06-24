package slapperx

import "time"

type OkBadCounter struct {
	Ok  counter
	Bad counter
}

type windowState int

const (
	Ready  windowState = 0
	Filled windowState = 1
)

// ring moving window buffer
type MovingWindow struct {
	counts [][]OkBadCounter
	state  []windowState
}

func newMovingWindow(nwindows int, buckets int) *MovingWindow {
	mw := &MovingWindow{}
	mw.state = make([]windowState, nwindows)

	mw.counts = make([][]OkBadCounter, nwindows)
	for i := 0; i < nwindows; i++ {
		mw.state[i] = Filled
		mw.counts[i] = make([]OkBadCounter, buckets)
	}
	return mw
}

func (mw *MovingWindow) getTimingsSlot(now time.Time) []OkBadCounter {
	n := int(now.UnixNano() / screenRefreshInterval.Nanoseconds())
	slot := n % len(mw.counts)
	if mw.state[slot] == Filled {
		oldSlot := (n - 1) % len(mw.counts)
		mw.state[oldSlot] = Filled
		mw.ResetSlot(slot)
	}
	return mw.counts[slot]
}

func (mw *MovingWindow) Reset() {
	for _, e := range mw.counts {
		for j := 0; j < len(e); j++ {
			e[j].Ok.Store(0)
			e[j].Bad.Store(0)
		}
	}
}

func (mw *MovingWindow) ResetSlot(slot int) {
	buckets := mw.counts[slot]
	for i := 0; i < len(buckets); i++ {
		buckets[i].Ok.Store(0)
		buckets[i].Bad.Store(0)
	}
	mw.state[slot] = Ready
}

// prepareHistogramData prepares data for histogram by aggregating OK and Bad requests
func (mw *MovingWindow) prepareHistogramData() ([]int64, []int64, int64) {
	tOk := make([]int64, len(mw.counts))
	tBad := make([]int64, len(mw.counts))

	max := int64(1)

	for i := 0; i < len(mw.counts); i++ {
		okBad := mw.counts[i]

		for j := 0; j < len(okBad); j++ {
			tOk[j] += okBad[j].Ok.Load()
			tBad[j] += okBad[j].Bad.Load()
			if sum := tOk[j] + tBad[j]; sum > max {
				max = sum
			}
		}
	}

	return tOk, tBad, max
}
