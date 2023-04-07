package main

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

	rateIncreaseStep = 50
	rateDecreaseStep = -50
)

var (
	desiredRate counter
	stats       Stats
	trgt        *Targeter
	ui          *UI
)

func main() {
	config := ParseFlags()
	fs := os.DirFS(".")
	requests := httpfile.HTTPFileParser(fs, config.Targets, true)
	if len(requests) == 0 {
		panic("No Requests")
	}

	trgt = NewTargeter(&requests, config.Timeout)
	//trgt.client.String()
	//os.Exit(0)

	ui = InitTerminal(config.MinY, config.MaxY)
	stats = Stats{}
	stats.initializeTimingsBucket(ui.buckets)

	quit := make(chan struct{}, 1)
	ticker, rateChanger := ticker(config.Rate, config.RampUp, quit)

	// start attackers
	trgt.Start(config.Workers, ticker, quit)

	// start reporter
	trgt.wg.Add(1)
	go func() {
		defer trgt.wg.Done()
		ui.reporter(quit)
	}()

	// blocking
	ui.keyPressListener(rateChanger)

	// bye
	close(quit)
	trgt.wg.Wait()
}
