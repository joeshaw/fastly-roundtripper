package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

type (
	requestContextKey  struct{}
	responseContextKey struct{}
)

type Transport struct {
	defaultBackend string
	backends       map[string]string

	Request func(fstreq *fsthttp.Request) error
}

func New(backend string) *Transport {
	return &Transport{
		defaultBackend: backend,
		backends:       map[string]string{},
	}
}

func (t *Transport) AddBackend(name, host string) {
	t.backends[strings.ToLower(host)] = name
}

func (t *Transport) getBackend(host string) string {
	backend, ok := t.backends[strings.ToLower(host)]
	if !ok {
		return t.defaultBackend
	}
	return backend
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	fstreq, err := fsthttp.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		return nil, err
	}
	fstreq.Header = fsthttp.Header(req.Header.Clone())

	if t.Request != nil {
		if err := t.Request(fstreq); err != nil {
			return nil, err
		}
	}

	fstresp, err := fstreq.Send(req.Context(), t.getBackend(req.URL.Host))
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(req.Context(), requestContextKey{}, fstreq)
	ctx = context.WithValue(ctx, responseContextKey{}, fstresp)

	resp := &http.Response{
		Request:    req.WithContext(ctx),
		StatusCode: fstresp.StatusCode,
		Header:     http.Header(fstresp.Header.Clone()),
		Body:       fstresp.Body,
	}

	return resp, nil
}

func FastlyRequestFromContext(ctx context.Context) *fsthttp.Request {
	fstreq, _ := ctx.Value(requestContextKey{}).(*fsthttp.Request)
	return fstreq
}

func FastlyResponseFromContext(ctx context.Context) *fsthttp.Response {
	fstresp, _ := ctx.Value(responseContextKey{}).(*fsthttp.Response)
	return fstresp
}
