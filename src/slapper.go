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
	quit := make(chan struct{}, 1)

	var logFile *LogFile = nil
	if config.LogFile != "" {
		logFile = NewLogFile(config.LogFile)
		defer logFile.Close()
	}

	trgt = NewTargeter(&requests, config.Timeout, logFile, config.Verbose)

	defer func() {
		close(quit)  // send all threads the quit signal
		trgt.Close() // wait and Close
	}()

	if !config.Verbose {
		ui = InitTerminal(config.MinY, config.MaxY)
		defer ui.Close()
	}

	stats = Stats{}
	if !config.Verbose {
		stats.initializeTimingsBucket(ui.lbc.buckets)
	}

	ticker := NewTicker(config.Rate)

	rampUpController := NewRamUpController(config.RampUp, config.Rate)
	go rampUpController.startRampUpTimeProcess(ticker.GetRateChanger())

	// start attackers
	var onTickChan = ticker.Start()
	defer ticker.Stop()

	trgt.Start(config.Workers, onTickChan)

	// blocking
	if !config.Verbose {
		ui.Show() // start Terminal output
	}
	keyPressListener(rampUpController.GetRateChanger())
}
