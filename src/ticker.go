package main

import "time"

func getDuration(rate uint64) (time.Duration, int) {
	var maxrate uint64 = 100
	multiplicator := rate / maxrate
	if multiplicator < 1 {
		multiplicator = 1
	}
	rate /= multiplicator

	d := time.Duration(1e9 / rate)
	return d, int(multiplicator)
}

func ticker(rate uint64, rampuptime time.Duration, quit <-chan struct{}) (<-chan time.Time, chan<- int64) {
	ticker := make(chan time.Time, 1)
	rateChanger := make(chan int64, 1)
	start := time.Now()

	go func() {
		for {
			now := time.Now()
			elapsed := now.Sub(start)
			if elapsed.Milliseconds() > rampuptime.Milliseconds() {
				rateChanger <- int64(rate)
				return
			}
			rateChanger <- (elapsed.Milliseconds() * int64(rate)) / rampuptime.Milliseconds()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// start main workers
	go func() {
		desiredRate.Store(int64(rate))
		duration, multiplicator := getDuration(rate)
		tck := time.NewTicker(duration)

		for {
			select {
			case r := <-rateChanger:
				tck.Stop()
				newRate := r
				desiredRate.Store(r)
				if newRate > 0 {
					duration, multiplicator = getDuration(uint64(newRate))
					tck = time.NewTicker(duration)
				} else {
					desiredRate.Store(0)
				}
			case t := <-tck.C:
				for i := 0; i < multiplicator; i++ {
					ticker <- t
				}
			case <-quit:
				return
			}
		}
	}()

	return ticker, rateChanger
}
