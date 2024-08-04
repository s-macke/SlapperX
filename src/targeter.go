package slapperx

import (
	"errors"
	"fmt"
	"github.com/s-macke/slapperx/src/tracing"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"
)

type Targeter struct {
	client   *tracing.Client
	wg       sync.WaitGroup
	idx      counter
	requests []http.Request
	logFile  *LogFile

	attackStartTime time.Time // time when the attack started

	verbose bool
}

func NewTargeter(requests *[]http.Request, timeout time.Duration, logFile *LogFile, verbose bool) *Targeter {
	client := tracing.NewTracingClient(timeout)

	trgt := &Targeter{
		client:   client,
		idx:      0,
		requests: *requests,
		logFile:  logFile,
		verbose:  verbose,
	}

	return trgt
}

func (trgt *Targeter) Close() {
	trgt.wg.Wait()
}

func (trgt *Targeter) nextRequest() *http.Request {
	idx := int(trgt.idx.Add(1))
	request := trgt.requests[idx%len(trgt.requests)]
	request.Body, _ = request.GetBody()
	return &request
}

func (trgt *Targeter) attack(client *tracing.Client, ch <-chan time.Time) {
	var dnsError *net.DNSError

	for {
		_, ok := <-ch
		if !ok { // channel closed
			return
		}
		request := trgt.nextRequest()
		stats.requestsSent.Add(1)

		start := time.Now()
		response, err := client.Do(request)
		if err != nil && trgt.verbose {
			fmt.Println("Error:", request.Method, request.URL, err)
		}
		if err == nil {
			_, err = io.ReadAll(response.Body)
			if err != nil && trgt.verbose {
				fmt.Println("Error:", request.Method, request.URL, err)
			}
			_ = response.Body.Close()
		}

		now := time.Now()
		elapsed := now.Sub(start)
		elapsedMs := float64(elapsed) / float64(time.Millisecond)

		stats.responsesReceived.Add(1)

		status := 0
		if err == nil {
			status = response.StatusCode
			stats.responses.status[status].Add(1)
		} else {
			switch {
			case
				errors.Is(err, io.EOF):
				stats.responses.ErrorEof.Add(1)
			case
				errors.Is(err, syscall.ECONNREFUSED):
				stats.responses.ErrorConnRefused.Add(1)
			case
				os.IsTimeout(err):
				stats.responses.ErrorTimeout.Add(1)
			case
				errors.As(err, &dnsError):
				stats.responses.ErrorNoSuchHost.Add(1)
			default:
				stats.responses.status[0].Add(1)
			}
		}

		if trgt.logFile != nil {
			trgt.logFile.WriteString(
				fmt.Sprintf("%s,%d,%d,%d,%d,%.1f\n",
					start.Format("2006-01-02T15:04:05.999999999"),
					start.Sub(trgt.attackStartTime).Milliseconds(),
					elapsed.Milliseconds(),
					status,
					stats.getInFlightRequests(),
					stats.currentSetRate))
		}

		if trgt.verbose {
			fmt.Println(request.Method, request.URL, status, elapsedMs)
			continue
		}
		// to test the latency distribution
		// elapsedMs = (math.Sin(elapsedMs)+1.1)*30. + math.Cos(float64(start.UnixMilli()/5000))*100 + 100.

		elapsedBucket := ui.lbc.calculateBucket(elapsedMs)
		timings := stats.timings.getTimingsSlot(now)
		if status >= 200 && status < 300 {
			timings[elapsedBucket].Ok.Add(1)
		} else {
			timings[elapsedBucket].Bad.Add(1)
		}
	}
}

func (trgt *Targeter) Start(workers uint, ticker <-chan time.Time) {
	trgt.attackStartTime = time.Now()
	// start attackers
	for i := uint(0); i < workers; i++ {
		trgt.wg.Add(1)
		go func() {
			defer trgt.wg.Done()
			trgt.attack(trgt.client, ticker)
		}()
	}
}
