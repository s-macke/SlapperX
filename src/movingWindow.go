package slapperx

import "time"

type OkBadCounter struct {
	Ok  int
	Bad int
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
	tOk      []int
	tBad     []int
}

func NewMovingWindow(nwindows int, nbuckets int) *MovingWindow {
	mw := &MovingWindow{
		nwindows: nwindows,
		nbuckets: nbuckets,
		tOk:      make([]int, nbuckets),
		tBad:     make([]int, nbuckets),
	}
	mw.state = make([]windowState, nwindows)

	mw.counts = make([][]OkBadCounter, nwindows)
	for i := 0; i < nwindows; i++ {
		mw.state[i] = Filled
		mw.counts[i] = make([]OkBadCounter, nbuckets)
	}
	return mw
}

type ResultStruct struct {
	elapsedMs int64
	status    int
	end       time.Time
}

func (mw *MovingWindow) Listen() chan ResultStruct {
	var resultChan = make(chan ResultStruct, 1000)

	go func() {
		for {
			select {
			case result := <-resultChan:
				elapsedBucket := ui.lbc.calculateBucket(float64(result.elapsedMs))
				slot := mw.getTimingsSlot(result.end) // end is basically now
				if result.status >= 200 && result.status < 300 {
					mw.counts[slot][elapsedBucket].Ok++
				} else {
					mw.counts[slot][elapsedBucket].Bad++
				}

			}
		}
	}()
	return resultChan
}

func (mw *MovingWindow) getTimingsSlot(now time.Time) int {
	n := int(now.UnixNano() / screenRefreshInterval.Nanoseconds())
	slot := n % len(mw.counts)
	if mw.state[slot] == Filled {
		oldSlot := (n - 1) % len(mw.counts)
		mw.state[oldSlot] = Filled
		mw.ResetSlot(slot)
	}
	return slot
}

func (mw *MovingWindow) Reset() {
	if mw.counts == nil {
		return
	}
	for _, e := range mw.counts {
		for j := 0; j < len(e); j++ {
			e[j].Ok = 0
			e[j].Bad = 0
		}
	}
}

func (mw *MovingWindow) ResetSlot(slot int) {
	buckets := mw.counts[slot]
	for i := 0; i < len(buckets); i++ {
		buckets[i].Ok = 0
		buckets[i].Bad = 0
	}
	mw.state[slot] = Ready
}

// prepareHistogramData prepares data for histogram by aggregating OK and Bad requests
func (mw *MovingWindow) prepareHistogramData() ([]int, []int, int) {
	for j := range mw.nbuckets {
		mw.tOk[j] = 0
		mw.tBad[j] = 0
	}

	maximum := 1

	for i := 0; i < mw.nwindows; i++ {
		okBad := mw.counts[i]

		for j := 0; j < mw.nbuckets; j++ {
			mw.tOk[j] += okBad[j].Ok
			mw.tBad[j] += okBad[j].Bad
			if sum := mw.tOk[j] + mw.tBad[j]; sum > maximum {
				maximum = sum
			}
		}
	}

	return mw.tOk, mw.tBad, maximum
}
