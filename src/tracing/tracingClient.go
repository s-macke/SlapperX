package tracing

import (
	"slapper/src/httpfile"
)

type Response struct {
	Status int
}

type TracingClient interface {
	Do(req *httpfile.Request, resp *Response) error
	String()
}
