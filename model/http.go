package model

import (
	"github.com/lauthrul/goutil/log"
	"github.com/valyala/fasthttp"
	"time"
)

var fc = &fasthttp.Client{}

func HttpDoTimeout(body []byte, method string, uri string, headers map[string]string, timeout time.Duration) ([]byte, int, error) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(uri)
	req.Header.SetMethod(method)

	switch method {
	case fasthttp.MethodPost:
		req.SetBody(body)
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// time.Second * 30
	err := fc.DoTimeout(req, resp, timeout)

	log.DebugF("%s -> [%d] %v\n", uri, resp.StatusCode(), err)

	return resp.Body(), resp.StatusCode(), err
}
