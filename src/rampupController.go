package slapperx

import (
	"time"
)

type RampUpController struct {
	startTime       time.Time
	rampUpTime      time.Duration
	maxRate         float64
	rateChangerChan chan float64
}

func NewRamUpController(rampUpTime time.Duration, maxRate float64) *RampUpController {
	r := &RampUpController{
		startTime:       time.Now(),
		rampUpTime:      rampUpTime,
		maxRate:         maxRate,
		rateChangerChan: make(chan float64),
	}
	go r.rateChangeListener()
	return r
}

func (r *RampUpController) rateChangeListener() {
	for {
		select {
		case rateChange := <-r.rateChangerChan:
			r.maxRate += rateChange
		}
	}
}

// StartRampUpProcess starts the ramp-up process.
func (r *RampUpController) startRampUpTimeProcess(rateChangerChan chan float64) {
	r.startTime = time.Now()
	lastRate := 0.
	for {
		now := time.Now()
		elapsed := now.Sub(r.startTime)
		if elapsed.Milliseconds() >= r.rampUpTime.Milliseconds() {
			maxRate := r.maxRate
			if maxRate != lastRate { // only send if rate has changed
				rateChangerChan <- maxRate
				lastRate = maxRate
			}
		} else {
			rateChangerChan <- (float64(elapsed.Milliseconds()) * r.maxRate) / float64(r.rampUpTime.Milliseconds())
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// GetRateChanger returns a channel for changing the ticker's rate during operation.
func (r *RampUpController) GetRateChanger() chan float64 {
	return r.rateChangerChan
}
