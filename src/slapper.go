package slapperx

import (
	"github.com/s-macke/slapperx/src/httpfile"
	"os"
	"time"
)

const (
	statsLines             = 3
	movingWindowsSize      = 10 // seconds
	screenRefreshFrequency = 5  // per second
	screenRefreshInterval  = time.Second / screenRefreshFrequency

	reservedWidthSpace  = 40
	reservedHeightSpace = 3

	rateIncreaseStep = 10
	rateDecreaseStep = -10
)

var (
	stats Stats
	trgt  *Targeter
	ui    *UI
)

func Main() {
	config := ParseFlags()

	fs := os.DirFS(".")
	requests := httpfile.HTTPFileParser(fs, config.Targets, true)
	if len(requests) == 0 {
		panic("No Requests")
	}

	trgt = NewTargeter(&requests, config.Timeout)
	ui = InitTerminal(config.MinY, config.MaxY)

	stats = Stats{}
	stats.initializeTimingsBucket(ui.buckets)

	quit := make(chan struct{}, 1)

	ticker := NewTicker(config.Rate)

	rampUpController := NewRamUpController(config.RampUp, config.Rate)
	go rampUpController.startRampUpTimeProcess(ticker.GetRateChanger())

	// start attackers
	var onTickChan = ticker.Start(quit)
	trgt.Start(config.Workers, onTickChan, quit)

	// start reporter
	trgt.wg.Add(1)
	go func() {
		defer trgt.wg.Done()
		ui.reporter(quit)
	}()

	// blocking
	ui.keyPressListener(rampUpController.GetRateChanger())

	// bye
	close(quit)
	trgt.Wait()
}
