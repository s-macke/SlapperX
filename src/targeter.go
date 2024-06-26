package slapperx

import (
	"bufio"
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

	file            *os.File
	fileWriter      *bufio.Writer
	attackStartTime time.Time

	verbose bool
}

func NewTargeter(requests *[]http.Request, timeout time.Duration, logFile string, verbose bool) *Targeter {
	client := tracing.NewTracingClient(timeout)

	trgt := &Targeter{
		client:   client,
		idx:      0,
		requests: *requests,
		file:     nil,
		verbose:  verbose,
	}

	if logFile != "" {
		var err error
		trgt.file, err = os.Create(logFile)
		if err != nil {
			panic(err)
		}
		trgt.fileWriter = bufio.NewWriterSize(trgt.file, 8192)
	}
	return trgt
}

func (trgt *Targeter) close() {
	if trgt.file != nil {
		err := trgt.fileWriter.Flush()
		if err != nil {
			panic(err)
		}
		err = trgt.file.Close()
		if err != nil {
			panic(err)
		}
	}
}

func (trgt *Targeter) nextRequest() *http.Request {
	idx := int(trgt.idx.Add(1))
	request := trgt.requests[idx%len(trgt.requests)]
	request.Body, _ = request.GetBody()
	return &request
}

func (trgt *Targeter) attack(client *tracing.Client, ch <-chan time.Time, quit <-chan struct{}) {

	var dnsError *net.DNSError

	for {
		select {
		case <-ch:
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

			if trgt.file != nil {
				_, err = trgt.fileWriter.WriteString(
					fmt.Sprintf("%s,%d,%d,%d\n",
						start.Format("2006-01-02T15:04:05.999999999"),
						start.Sub(trgt.attackStartTime).Milliseconds(),
						elapsed.Milliseconds(),
						status))
				if err != nil {
					panic(err)
				}
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

		case <-quit:
			return
		}
	}
}

func (trgt *Targeter) Start(workers uint, ticker <-chan time.Time, quit <-chan struct{}) {
	trgt.attackStartTime = time.Now()
	// start attackers
	for i := uint(0); i < workers; i++ {
		trgt.wg.Add(1)
		go func() {
			defer trgt.wg.Done()
			trgt.attack(trgt.client, ticker, quit)
		}()
	}
}

func (trgt *Targeter) Wait() {
	trgt.wg.Wait()
}
