package tracing

import (
	"crypto/tls"
	"fmt"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"time"
)

type TracingFastHttpClient struct {
	client             *fasthttp.Client
	dialer             *fasthttp.TCPDialer
	hc                 map[string]*fasthttp.HostClient
	currentConnections int
	openedConnections  int
	closedConnections  int
}

func NewFastHttpTracingClient(timeout time.Duration) *TracingFastHttpClient {
	dial := &fasthttp.TCPDialer{Concurrency: 1000}
	tracingClient := &TracingFastHttpClient{
		dialer: dial,
		hc:     nil,
	}
	tracingClient.client = &fasthttp.Client{
		Name:                     "", // default ist fasthttp
		NoDefaultUserAgentHeader: true,
		Dial:                     tracingClient.Dial,
		DialDualStack:            false, // for ipv4 and ipv6

		TLSConfig: &tls.Config{InsecureSkipVerify: true},
		//MaxIdleConnDuration:           0,
		//MaxConnDuration:               0,
		//MaxIdemponentCallAttempts:     0,
		//ReadBufferSize:                0,
		//WriteBufferSize:               0,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		/*
			MaxResponseBodySize:           0,
			DisableHeaderNamesNormalizing: false,
			DisablePathNormalizing:        false,
			MaxConnWaitTimeout:            0,
			RetryIf:                       nil,
			ConnPoolStrategy:              0,
		*/
		ConfigureClient: func(hc *fasthttp.HostClient) error {
			tracingClient.hc[hc.Addr] = hc
			return nil
		},
	}
	tracingClient.String()
	os.Exit(1)
	return tracingClient
}

func (t *TracingFastHttpClient) Dial(addr string) (net.Conn, error) {
	conn, err := t.dialer.Dial(addr)
	c := TracingConnection{Conn: conn}
	c.OnEventCallback = func(clientClosed bool, serverClosed bool, err error) {
		t.closedConnections++
		t.currentConnections--
	}

	if err == nil {
		t.openedConnections++
		t.currentConnections++
	}
	return c, err
}

func (t *TracingFastHttpClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return t.client.Do(req, resp)
}

func (t *TracingFastHttpClient) String() {
	fmt.Println("Current Connections:", t.currentConnections)
	fmt.Println("Opened Connections:", t.openedConnections)
	fmt.Println("Closed Connections:", t.closedConnections)
	fmt.Println("MaxConnsPerHost", t.client.MaxConnsPerHost)
	fmt.Println("MaxIdleConnDuration", t.client.MaxIdleConnDuration)
	fmt.Println("MaxConnDuration", t.client.MaxConnDuration)
	fmt.Println("MaxIdemponentCallAttempts", t.client.MaxIdemponentCallAttempts)
	fmt.Println("ReadBufferSize", t.client.ReadBufferSize) // 4096, also used for header parsing
	fmt.Println("WriteBufferSize", t.client.WriteBufferSize)
	fmt.Println("ReadTimeout", t.client.ReadTimeout)
	fmt.Println("WriteTimeout", t.client.WriteTimeout)
	fmt.Println("MaxResponseBodySize", t.client.MaxResponseBodySize)
	fmt.Println("DisableHeaderNamesNormalizing", t.client.DisableHeaderNamesNormalizing)
	fmt.Println("DisablePathNormalizing", t.client.DisablePathNormalizing)
	fmt.Println("MaxConnWaitTimeout", t.client.MaxConnWaitTimeout)
	fmt.Println("RetryIf", t.client.RetryIf)
	fmt.Println("ConnPoolStrategy", t.client.ConnPoolStrategy)
	for key, value := range t.hc {
		//fmt.Println(key, value.ConnsCount())
		fmt.Println(key, value.PendingRequests())
		//defaultReadBufferSize  = 4096
		//defaultWriteBufferSize = 4096

	}
	fmt.Println("")
}
