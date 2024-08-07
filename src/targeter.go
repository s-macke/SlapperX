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

type AttackResponse struct {
	status int
	err    error
	start  time.Time
	end    time.Time
	body   []byte
}

func (trgt *Targeter) DoRequest(request *http.Request, doStoreBody bool) AttackResponse {
	attackResponse := AttackResponse{
		status: 0,
		err:    nil,
	}
	attackResponse.start = time.Now()

	response, err := trgt.client.Do(request)
	if err != nil && trgt.verbose {
		fmt.Println("Error:", request.Method, request.URL, err)
	}
	if err == nil {
		attackResponse.status = response.StatusCode
		if doStoreBody {
			attackResponse.body, err = io.ReadAll(response.Body)
		} else {
			_, err = io.ReadAll(response.Body)
		}
		if err != nil && trgt.verbose {
			fmt.Println("Error:", request.Method, request.URL, err)
		}
		_ = response.Body.Close()
	}
	attackResponse.err = err
	attackResponse.end = time.Now()
	return attackResponse
}

func FillStats(request *http.Request, response AttackResponse,
	currentSetRate float64, currentInFlightRequests int64) {
	var dnsError *net.DNSError
	stats.responsesReceived.Add(1)
	if response.err == nil {
		stats.responses.status[response.status].Add(1)
	} else {
		switch {
		case
			errors.Is(response.err, io.EOF):
			stats.responses.ErrorEof.Add(1)
		case
			errors.Is(response.err, syscall.ECONNREFUSED):
			stats.responses.ErrorConnRefused.Add(1)
		case
			os.IsTimeout(response.err):
			stats.responses.ErrorTimeout.Add(1)
		case
			errors.As(response.err, &dnsError):
			stats.responses.ErrorNoSuchHost.Add(1)
		default:
			stats.responses.status[0].Add(1)
		}
	}

	elapsed := response.end.Sub(response.start)
	elapsedMs := float64(elapsed) / float64(time.Millisecond)
	// to test the latency distribution
	// elapsedMs = (math.Sin(elapsedMs)+1.1)*30. + math.Cos(float64(start.UnixMilli()/5000))*100 + 100.

	if trgt.logFile != nil {
		trgt.logFile.WriteString(
			fmt.Sprintf("%s,%d,%d,%d,%d,%.1f\n",
				response.start.Format("2006-01-02T15:04:05.999999999"),
				response.start.Sub(trgt.attackStartTime).Milliseconds(),
				elapsed.Milliseconds(),
				response.status,
				currentInFlightRequests,
				currentSetRate))
	}

	if trgt.verbose {
		fmt.Println(request.Method, request.URL, response.status, elapsedMs)
		return
	}

	elapsedBucket := ui.lbc.calculateBucket(elapsedMs)
	timings := stats.timings.getTimingsSlot(response.end) // end is basically now
	if response.status >= 200 && response.status < 300 {
		timings[elapsedBucket].Ok.Add(1)
	} else {
		timings[elapsedBucket].Bad.Add(1)
	}
}

func (trgt *Targeter) attack(ch <-chan time.Time) {
	for {
		_, ok := <-ch
		if !ok { // channel closed
			return
		}
		request := trgt.nextRequest()
		stats.requestsSent.Add(1)

		// Save the rate when the request started
		currentSetRate := stats.currentSetRate
		currentInFlightRequests := stats.getInFlightRequests()

		response := trgt.DoRequest(request, false)
		FillStats(request, response, currentSetRate, currentInFlightRequests)
	}
}

func (trgt *Targeter) Start(workers uint, ticker <-chan time.Time) {
	trgt.attackStartTime = time.Now()
	// start attackers
	for i := uint(0); i < workers; i++ {
		trgt.wg.Add(1)
		go func() {
			defer trgt.wg.Done()
			trgt.attack(ticker)
		}()
	}
}
