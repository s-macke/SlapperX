package slapperx

import (
	"fmt"
	"github.com/s-macke/slapperx/src/httpfile"
	"os"
	"time"
)

const (
	movingWindowsSize      = 10 // seconds
	screenRefreshFrequency = 5  // per second
	screenRefreshInterval  = time.Second / screenRefreshFrequency
)

var (
	stats Stats
	ui    *UI
	trgt  *Targeter
)

func Main() {
	config := ParseFlags()

	requests, err := httpfile.HTTPFileParser(config.Targets, config.Overrides, true)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse HTTP file: %v\n", err)
		return
	}
	if len(requests) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No requests found in the HTTP file\n")
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

	stats = Stats{}

	var resultChan chan ResultStruct = nil
	if !config.Verbose {
		ui = InitTerminal(config.MinY, config.MaxY)
		defer ui.Close()
		stats.initializeTimingsBucket(ui.lbc.buckets)
		resultChan = stats.timings.Listen()
	}

	trgt = NewTargeter(&requests, config.Timeout, logFile, config.Verbose, resultChan)

	defer func() {
		close(quit)  // send all threads the quit signal
		trgt.Close() // wait and Close
	}()

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

	// Create and start keyboard handler
	keyboard := InitKeyboard(rampUpController)
	keyboard.Start()
}
