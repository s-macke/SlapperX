package main

import (
	"crypto/tls"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
)

type targeter struct {
	client   *http.Client
	wg       sync.WaitGroup
	idx      counter
	requests []http.Request
}

func NewTargeter(requests *[]http.Request, timeout time.Duration) *targeter {
	tr := &http.Transport{
		DisableKeepAlives:   false,
		DisableCompression:  true,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return &targeter{
		client:   client,
		idx:      0,
		requests: *requests,
	}
}

func (trgt *targeter) nextRequest() *http.Request {
	idx := int(trgt.idx.Add(1))
	return &trgt.requests[idx%len(trgt.requests)]
}

func (trgt *targeter) attack(client *http.Client, ch <-chan time.Time, quit <-chan struct{}) {

	for {
		select {
		case <-ch:
			request := trgt.nextRequest()
			stats.requestsSent.Add(1)

			start := time.Now()
			response, err := client.Do(request)
			if err == nil {
				_, err = io.ReadAll(response.Body)
				err = response.Body.Close()
			}
			now := time.Now()

			elapsed := now.Sub(start)
			elapsedMs := float64(elapsed) / float64(time.Millisecond)
			correctedElapsedMs := elapsedMs - ui.startMs
			elapsedBucket := int(math.Log(correctedElapsedMs) / math.Log(ui.logBase))

			// first bucket is for requests faster then minY,
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

func (trgt *targeter) Start(workers uint, ticker <-chan time.Time, quit <-chan struct{}) {
	// start attackers
	for i := uint(0); i < workers; i++ {
		trgt.wg.Add(1)
		go func() {
			defer trgt.wg.Done()
			trgt.attack(trgt.client, ticker, quit)
		}()
	}

}
