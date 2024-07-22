package slapperx

type StatsResponse struct {
	status           [1024]counter
	ErrorEof         counter
	ErrorTimeout     counter
	ErrorConnRefused counter
	ErrorNoSuchHost  counter
}

type Stats struct {
	currentSetRate    float64
	requestsSent      counter
	responsesReceived counter

	responses StatsResponse

	// ring moving window buffer
	timings *MovingWindow
}

func (s *Stats) reset() {
	s.requestsSent.Store(0)
	s.responsesReceived.Store(0)

	s.timings.Reset()

	for i := 0; i < len(s.responses.status); i++ {
		s.responses.status[i].Store(0)
	}
}

func (s *Stats) initializeTimingsBucket(buckets int) {
	s.timings = newMovingWindow(movingWindowsSize*screenRefreshFrequency, buckets)
}
