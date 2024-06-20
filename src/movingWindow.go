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
	counts   [][]OkBadCounter
	state    []windowState
	nwindows int
	nbuckets int
	tOk      []int64
	tBad     []int64
}

func newMovingWindow(nwindows int, nbuckets int) *MovingWindow {
	mw := &MovingWindow{
		nwindows: nwindows,
		nbuckets: nbuckets,
		tOk:      make([]int64, nbuckets),
		tBad:     make([]int64, nbuckets),
	}
	mw.state = make([]windowState, nwindows)

	mw.counts = make([][]OkBadCounter, nwindows)
	for i := 0; i < nwindows; i++ {
		mw.state[i] = Filled
		mw.counts[i] = make([]OkBadCounter, nbuckets)
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
	for j := range mw.nbuckets {
		mw.tOk[j] = 0
		mw.tBad[j] = 0
	}

	maximum := int64(1)

	for i := 0; i < mw.nwindows; i++ {
		okBad := mw.counts[i]

		for j := 0; j < mw.nbuckets; j++ {
			mw.tOk[j] += okBad[j].Ok.Load()
			mw.tBad[j] += okBad[j].Bad.Load()
			if sum := mw.tOk[j] + mw.tBad[j]; sum > maximum {
				maximum = sum
			}
		}
	}

	return mw.tOk, mw.tBad, maximum
}
