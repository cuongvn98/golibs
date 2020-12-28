// Package instrument provides ...
package instrument

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// HTTPReqProperties - are the metrics properties for the metrics based on  client request
type HTTPReqProperties struct {
	ServiceName string
	Handler     string
	Method      string
	Code        string
}

// Recorder - know how to record and measure request
type Recorder interface {
	ObserveHTTPRequestDuration(ctx context.Context, props HTTPReqProperties, duration time.Duration)
	ObserveHTTPRequestCount(ctx context.Context, props HTTPReqProperties)
	IncCurrentRequest(ctx context.Context, props HTTPReqProperties)
	DescCurrentRequest(ctx context.Context, props HTTPReqProperties)
}

type middleware struct {
	serviceName string
	next        http.Handler
	recorder    Recorder
}

// New - create instrument middleware
func New(next http.Handler, recorder Recorder, serviceName string) http.Handler {
	return &middleware{next: next, recorder: recorder, serviceName: serviceName}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := newWrapResponseWriter(w)
	start := time.Now()
	props := HTTPReqProperties{
		ServiceName: m.serviceName,
		Handler:     r.URL.Path,
		Method:      r.Method,
	}
	m.recorder.IncCurrentRequest(r.Context(), props)
	defer func() {
		code := strconv.Itoa(rw.Status())
		props.Code = code
		m.recorder.ObserveHTTPRequestDuration(r.Context(), props, time.Since(start))
		m.recorder.ObserveHTTPRequestCount(r.Context(), props)
		m.recorder.DescCurrentRequest(r.Context(), props)
	}()
	m.next.ServeHTTP(rw, r)
}
