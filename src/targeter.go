package slapperx

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/s-macke/slapperx/src/tracing"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"
)

type Targeter struct {
	client   *tracing.TracingClient
	wg       sync.WaitGroup
	idx      counter
	requests []http.Request

	file            *os.File
	fileWriter      *bufio.Writer
	attackStartTime time.Time
}

func NewTargeter(requests *[]http.Request, timeout time.Duration, logFile string) *Targeter {
	client := tracing.NewTracingClient(timeout)

	trgt := &Targeter{
		client:   client,
		idx:      0,
		requests: *requests,
		file:     nil,
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
	return &trgt.requests[idx%len(trgt.requests)]
}

func (trgt *Targeter) attack(client *tracing.TracingClient, ch <-chan time.Time, quit <-chan struct{}) {

	for {
		select {
		case <-ch:
			request := trgt.nextRequest()
			stats.requestsSent.Add(1)

			start := time.Now()
			response, err := client.Do(request)
			if err == nil {
				_, err = io.ReadAll(response.Body)
				_ = response.Body.Close()
			}
			now := time.Now()

			elapsed := now.Sub(start)
			elapsedMs := float64(elapsed) / float64(time.Millisecond)
			correctedElapsedMs := elapsedMs - ui.startMs
			elapsedBucket := int(math.Log(correctedElapsedMs) / math.Log(ui.logBase))

			// first bucket is for requests faster than minY,
			// last of for ones slower then maxY
			if elapsedBucket < 0 {
				elapsedBucket = 0
			} else if elapsedBucket >= int(ui.buckets)-1 {
				elapsedBucket = int(ui.buckets) - 1
			} else {
				elapsedBucket = elapsedBucket + 1
			}

			stats.responsesReceived.Add(1)

			status := 0
			if err == nil {
				status = response.StatusCode
				stats.responses[status].Add(1)
			} else {
				switch {
				case
					errors.Is(err, io.EOF):
					stats.responsesErrorEof.Add(1)
				case
					errors.Is(err, syscall.ECONNREFUSED):
					stats.responsesErrorConnRefused.Add(1)
				case
					os.IsTimeout(err):
					stats.responsesErrorTimeout.Add(1)
				default:
					stats.responses[0].Add(1)
				}

			}

			tOk, tBad := stats.getTimingsSlot(now)
			if status >= 200 && status < 300 {
				tOk[elapsedBucket].Add(1)
			} else {
				tBad[elapsedBucket].Add(1)
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
