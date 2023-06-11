package slapperx

import (
	"time"
)

type RampUpController struct {
	startTime       time.Time
	rampUpTime      time.Duration
	maxRate         counter
	rateChangerChan chan int64
}

func NewRamUpController(rampUpTime time.Duration, maxRate int64) *RampUpController {
	r := &RampUpController{
		startTime:       time.Now(),
		rampUpTime:      rampUpTime,
		maxRate:         counter(maxRate),
		rateChangerChan: make(chan int64),
	}
	go r.rateChangeListener()
	return r
}

func (r *RampUpController) rateChangeListener() {
	for {
		select {
		case rateChange := <-r.rateChangerChan:
			r.maxRate.Add(rateChange)
		}
	}
}

// StartRampUpProcess starts the ramp-up process.
func (r *RampUpController) startRampUpTimeProcess(rateChangerChan chan int64) {
	r.startTime = time.Now()
	lastRate := int64(0)
	for {
		now := time.Now()
		elapsed := now.Sub(r.startTime)
		if elapsed.Milliseconds() >= r.rampUpTime.Milliseconds() {
			maxRate := r.maxRate.Load()
			if maxRate != lastRate { // only send if rate has changed
				rateChangerChan <- maxRate
				lastRate = maxRate
			}
		} else {
			rateChangerChan <- (elapsed.Milliseconds() * r.maxRate.Load()) / r.rampUpTime.Milliseconds()
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// GetRateChanger returns a channel for changing the ticker's rate during operation.
func (r *RampUpController) GetRateChanger() chan int64 {
	return r.rateChangerChan
}
