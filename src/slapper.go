package slapperx

import (
	"fmt"
	"github.com/s-macke/slapperx/src/httpfile"
	"net/http"
	"os"
	"strings"
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
)

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string // Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)                             // Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	} // Return the request as a string
	return strings.Join(request, "\n")
}

func Main() {
	var trgt *Targeter

	config := ParseFlags()

	fs := os.DirFS(".")
	requests := httpfile.HTTPFileParser(fs, config.Targets, true)
	if len(requests) == 0 {
		panic("No Requests")
	}

	trgt = NewTargeter(&requests, config.Timeout, config.LogFile)
	defer trgt.close()

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
