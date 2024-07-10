package slapperx

import (
	"fmt"
	"github.com/s-macke/slapperx/src/httpfile"
	"time"
)

const (
	movingWindowsSize      = 10 // seconds
	screenRefreshFrequency = 5  // per second
	screenRefreshInterval  = time.Second / screenRefreshFrequency

	rateIncreaseStep = 10
	rateDecreaseStep = -10
)

var (
	stats Stats
	ui    *UI
	trgt  *Targeter
)

func Main() {
	config := ParseFlags()

	requests := httpfile.HTTPFileParser(config.Targets, config.Overrides, true)
	if len(requests) == 0 {
		panic("No Requests")
	}
	if config.Verbose {
		fmt.Println("Requests:", len(requests))
	}

	trgt = NewTargeter(&requests, config.Timeout, config.LogFile, config.Verbose)
	defer trgt.close()
	if !config.Verbose {
		ui = InitTerminal(config.MinY, config.MaxY)
		defer ui.close()
	}

	stats = Stats{}
	if !config.Verbose {
		stats.initializeTimingsBucket(ui.lbc.buckets)
	}

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
		if !config.Verbose {
			ui.reporter(quit)
		}
	}()

	// blocking
	if config.Verbose {
		<-make(chan bool) // just wait for Ctrl-C
	} else {
		ui.keyPressListener(rampUpController.GetRateChanger())
	}

	// bye
	close(quit)
	trgt.Wait()
}
