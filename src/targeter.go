package main

import (
	"github.com/valyala/fasthttp"
	"math"
	"sync"
	"time"
)

type Targeter struct {
	//client   *tracing.TracingClient
	client   *fasthttp.Client
	wg       sync.WaitGroup
	idx      counter
	requests []fasthttp.Request
}

func NewTargeter(requests *[]fasthttp.Request, timeout time.Duration) *Targeter {
	//client := tracing.NewTracingClient(timeout)
	client := &fasthttp.Client{
		//Name:                          "",
		//NoDefaultUserAgentHeader:      false,
		//Dial:                          nil,
		//DialDualStack:                 false,
		//TLSConfig:                     nil,
		MaxConnsPerHost: math.MaxInt,
		//MaxIdleConnDuration:           0,
		//MaxConnDuration:               0,
		//MaxIdemponentCallAttempts:     0,
		//ReadBufferSize:                0,
		//WriteBufferSize:               0,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		//MaxResponseBodySize:           0,
		//DisableHeaderNamesNormalizing: false,
		//DisablePathNormalizing:        false,
		//MaxConnWaitTimeout:            0,
		//RetryIf:                       nil,
		//ConnPoolStrategy:              0,
		//ConfigureClient:               nil,
	}

	return &Targeter{
		client:   client,
		idx:      0,
		requests: *requests,
	}
}

func (trgt *Targeter) nextRequest() *fasthttp.Request {
	idx := int(trgt.idx.Add(1))
	return &trgt.requests[idx%len(trgt.requests)]
}

func (trgt *Targeter) attack(client *fasthttp.Client, ch <-chan time.Time, quit <-chan struct{}) {

	for {
		select {
		case <-ch:
			request := trgt.nextRequest()
			stats.requestsSent.Add(1)

			start := time.Now()
			var response = fasthttp.AcquireResponse()
			err := client.Do(request, response)
			if err == nil {
				_ = response.Body()
			}
			statusCode := response.StatusCode()
			fasthttp.ReleaseResponse(response)
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
				status = statusCode
			}

			stats.responses[status].Add(1)
			tOk, tBad := stats.getTimingsSlot(now)
			if status >= 200 && status < 300 {
				tOk[elapsedBucket].Add(1)
			} else {
				tBad[elapsedBucket].Add(1)
			}
		case <-quit:
			return
		}
	}
}

func (trgt *Targeter) Start(workers uint, ticker <-chan time.Time, quit <-chan struct{}) {
	// start attackers
	for i := uint(0); i < workers; i++ {
		trgt.wg.Add(1)
		go func() {
			defer trgt.wg.Done()
			trgt.attack(trgt.client, ticker, quit)
		}()
	}

}
