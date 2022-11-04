package tracing

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type TracingClient struct {
	transport          *http.Transport
	dialer             net.Dialer
	client             http.Client
	currentConnections int
	openedConnections  int
	closedConnections  int
}

func NewTracingClient(timeout time.Duration) *TracingClient {
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

	tc := &TracingClient{
		transport:          transport,
		dialer:             dial,
		client:             client,
		currentConnections: 0,
		openedConnections:  0,
		closedConnections:  0,
	}

	tc.transport.DialContext = tc.DialContext

	return tc
}

func (t *TracingClient) String() {
	fmt.Println("Current Connections:", t.currentConnections)
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

func (t *TracingClient) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := t.dialer.DialContext(ctx, network, address)
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

func (t *TracingClient) Do(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.client.Do(req)
	return
}

func call(tracingClient *TracingClient) {
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
		tracingClient.currentConnections,
		tracingClient.openedConnections,
		tracingClient.closedConnections,
	)
	//fmt.Println(client.Transport)
}
