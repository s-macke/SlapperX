package tracing

import (
	"github.com/valyala/fasthttp"
	"net/http"
)

type Request struct {
	req1 *fasthttp.Request
	req2 *http.Request
}

type TracingClient interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
	String()
}
