package slapperx

import (
	"time"
)

type Ticker struct {
	rate            int64
	multiplier      int64
	tickDuration    time.Duration
	rateChangerChan chan int64
}

// NewTicker creates a new ticker instance with a given rate and ramp-up time.
func NewTicker(rate int64) *Ticker {
	t := &Ticker{
		rateChangerChan: make(chan int64),
	}
	t.setTickDuration(rate)
	return t
}

// setTickDuration sets the duration between ticks based on the given rate.
func (t *Ticker) setTickDuration(rate int64) {
	const maxTickRate int64 = 100

	t.rate = rate
	t.multiplier = rate / maxTickRate
	if t.multiplier < 1 {
		t.multiplier = 1
	}
	t.rate /= t.multiplier
	t.tickDuration = time.Duration(1e9 / t.rate)
}

// GetRateChanger returns a channel for changing the ticker's rate during operation.
func (t *Ticker) GetRateChanger() chan int64 {
	return t.rateChangerChan
}

// Start initializes the tick process and returns a channel to receive tick events.
func (t *Ticker) Start(quit <-chan struct{}) <-chan time.Time {
	ticker := make(chan time.Time)

	// start main workers
	go func() {
		stats.currentSetRate.Store(t.rate)
		tck := time.NewTicker(t.tickDuration)

		for {
			select {
			case newRate := <-t.rateChangerChan:
				stats.currentSetRate.Store(newRate)
				if newRate > 0 {
					t.setTickDuration(newRate)
					tck.Reset(t.tickDuration)
				} else {
					stats.currentSetRate.Store(newRate)
				}

			case onTick := <-tck.C:
				for i := int64(0); i < t.multiplier; i++ {
					ticker <- onTick
				}

			case <-quit:
				return
			}
		}
	}()
	return ticker
}
