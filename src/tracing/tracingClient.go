package tracing

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

type Client struct {
	transport          *http.Transport
	dialer             net.Dialer
	client             http.Client
	CurrentConnections int32
	openedConnections  int32
	closedConnections  int32
}

func NewTracingClient(timeout time.Duration) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	transport.MaxIdleConns = 0 // No Limit
	transport.MaxIdleConnsPerHost = 100

	dial := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  net.DefaultResolver,
	}

	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	tc := &Client{
		transport:          transport,
		dialer:             dial,
		client:             client,
		CurrentConnections: 0,
		openedConnections:  0,
		closedConnections:  0,
	}

	tc.transport.DialContext = tc.DialContext

	return tc
}

func (t *Client) String() {
	//fmt.Println("\033[H") // clean screen
	fmt.Println("\033[2J")
	fmt.Println("Current Connections:", t.CurrentConnections)
	fmt.Println("Opened Connections:", t.openedConnections)
	fmt.Println("Closed Connections:", t.closedConnections)
	fmt.Println("Force Attempt HTTP2:", t.transport.ForceAttemptHTTP2)
	fmt.Println("Max Idle Connections:", t.transport.MaxIdleConns)
	fmt.Println("Max Idle Connections Per Host:", t.transport.MaxIdleConnsPerHost)
	fmt.Println("Max Response Header Bytes:", t.transport.MaxResponseHeaderBytes)
	fmt.Println("TLS Handshake Timeout:", t.transport.TLSHandshakeTimeout)
	fmt.Println("Expect Continue Timeout:", t.transport.ExpectContinueTimeout)
	fmt.Println("Idle Connection Timeout:", t.transport.IdleConnTimeout)
	fmt.Println("Response Header Timeout:", t.transport.ResponseHeaderTimeout)
	fmt.Println("TLS Next Protocol:", t.transport.TLSNextProto)
	//fmt.Println("TLS Client Config:", t.transport.TLSClientConfig)
	fmt.Println("TLS Insecure Skip Verify:", t.transport.TLSClientConfig.InsecureSkipVerify)
	//fmt.Println("TLS Root CAs:", t.transport.TLSClientConfig.RootCAs)
	fmt.Println("Disable Compression:", t.transport.DisableCompression)
	fmt.Println("Disable Keep Alives:", t.transport.DisableKeepAlives)
	fmt.Println("Max Connections Per Host:", t.transport.MaxConnsPerHost)
	fmt.Println("Read Buffer Size:", t.transport.ReadBufferSize)
	fmt.Println("Write Buffer Size:", t.transport.WriteBufferSize)
	fmt.Println("Dialer Timeout:", t.dialer.Timeout)
	fmt.Println("Dialer Deadline:", t.dialer.Deadline)
	fmt.Println("Dialer Keep Alive:", t.dialer.KeepAlive)
	fmt.Println("Dialer Fallback Delay:", t.dialer.FallbackDelay)
	fmt.Println("Dialer Fallback Local Address:", t.dialer.LocalAddr)
	fmt.Println("Resolver Strict Errors:", t.dialer.Resolver.StrictErrors)
	fmt.Println("Resolver Prefer Go:", t.dialer.Resolver.PreferGo)
	fmt.Println("Client Timeout:", t.client.Timeout)
	fmt.Println("Client Cookie Jar:", t.client.Jar)
}

func (t *Client) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := t.dialer.DialContext(ctx, network, address)
	c := &Connection{Conn: conn}
	c.OnEventCallback = func(clientClosed bool, serverClosed bool, err error) {
		atomic.AddInt32(&t.closedConnections, 1)
		atomic.AddInt32(&t.CurrentConnections, -1)
	}

	if err == nil {
		atomic.AddInt32(&t.openedConnections, 1)
		atomic.AddInt32(&t.CurrentConnections, 1)
	}
	return c, err
}

func (t *Client) Do(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.client.Do(req)
	return
}

func call(tracingClient *Client) {
	req, err := http.NewRequest("GET", "http://localhost:8080/hello/", nil)
	req.Header.Set("Connection", "keep-alive")

	if err != nil {
		fmt.Print(err.Error())
	}
	start := time.Now().UnixNano()
	resp, err := tracingClient.client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err.Error())
		}
	}(resp.Body)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	end := time.Now().UnixNano()
	fmt.Println(start/1e6, end/1e6, (end-start)/1e6, resp.Status,
		tracingClient.CurrentConnections,
		tracingClient.openedConnections,
		tracingClient.closedConnections,
	)
	//fmt.Println(client.Transport)
}
