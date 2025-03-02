package slapperx

import (
	"math"
	"time"
)

// RampUpController constants
const (
	rateIncreaseStep = 10
	rateDecreaseStep = -10
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
			r.maxRate = math.Max(0.0001, r.maxRate)
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

// ChangeRate allows direct modification of the rate by a delta amount
func (r *RampUpController) ChangeRate(delta float64) {
	r.rateChangerChan <- delta
}

// IncreaseRate increases the rate by the standard step
func (r *RampUpController) IncreaseRate() {
	r.ChangeRate(rateIncreaseStep)
}

// DecreaseRate decreases the rate by the standard step
func (r *RampUpController) DecreaseRate() {
	r.ChangeRate(rateDecreaseStep)
}

// SetRate sets the rate to an absolute value (implemented as a delta from current)
func (r *RampUpController) SetRate(newRate float64) {
	// Calculate the delta needed to reach newRate
	// This needs to be sent through the channel to ensure thread safety
	delta := newRate - r.maxRate
	r.ChangeRate(delta)
}

// GetRateChanger returns a channel for changing the ticker's rate during operation.
// This is kept for compatibility but new code should use the ChangeRate/SetRate methods.
func (r *RampUpController) GetRateChanger() chan float64 {
	return r.rateChangerChan
}
