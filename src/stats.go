package slapperx

import (
	"time"
)

type StatsResponse struct {
	status           [1024]counter
	ErrorEof         counter
	ErrorTimeout     counter
	ErrorConnRefused counter
	ErrorNoSuchHost  counter
}

type Stats struct {
	currentSetRate    counter
	requestsSent      counter
	responsesReceived counter

	responses StatsResponse

	// ring buffer
	timingsOk  [][]counter
	timingsBad [][]counter
}

func (s *Stats) getTimingsSlot(now time.Time) ([]counter, []counter) {
	//n := int(now.UnixNano() / 100000000)
	n := int(now.UnixNano() / screenRefreshInterval.Nanoseconds())

	slot := n % len(s.timingsOk)
	return s.timingsOk[slot], s.timingsBad[slot]
}

func resetStats() {
	s := Stats{}
	s.requestsSent.Store(0)
	s.responsesReceived.Store(0)

	for _, ok := range s.timingsOk {
		for i := 0; i < len(ok); i++ {
			ok[i].Store(0)
		}
	}

	for _, bad := range s.timingsBad {
		for i := 0; i < len(bad); i++ {
			bad[i].Store(0)
		}
	}

	for i := 0; i < len(s.responses.status); i++ {
		s.responses.status[i].Store(0)
	}
}

func (s *Stats) initializeTimingsBucket(buckets int) {
	s.timingsOk = make([][]counter, movingWindowsSize*screenRefreshFrequency)
	for i := 0; i < len(s.timingsOk); i++ {
		s.timingsOk[i] = make([]counter, buckets)
	}

	s.timingsBad = make([][]counter, movingWindowsSize*screenRefreshFrequency)
	for i := 0; i < len(s.timingsBad); i++ {
		s.timingsBad[i] = make([]counter, buckets)
	}

	go func() {
		for now := range time.Tick(screenRefreshInterval) {
			// clean next timing slot which is last one in ring buffer
			next := now.Add(screenRefreshInterval)
			tOk, tBad := s.getTimingsSlot(next)
			for i := 0; i < len(tOk); i++ {
				tOk[i].Store(0)
			}

			for i := 0; i < len(tBad); i++ {
				tBad[i].Store(0)
			}
		}
	}()
}
