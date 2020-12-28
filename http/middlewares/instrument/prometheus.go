// Package instrument provides ...
package instrument

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type prometheusRecorder struct {
	requestLatency *prometheus.HistogramVec
	requestCounter *prometheus.CounterVec
	requestCurrent *prometheus.GaugeVec
}

// PrometheusConfig - configuraion for prometheus recorder
type PrometheusConfig struct {
	// Prefix is the prefix that will be set on the metrics, by default it will be empty.
	Prefix string
	// DurationBuckets are the buckets used by Prometheus for the HTTP request duration metrics,
	// by default uses Prometheus default buckets (from 5ms to 10s).
	DurationBuckets []float64
	// Registry is the registry that will be used by the recorder to store the metrics,
	// if the default registry is not used then it will use the default one.
	Registry prometheus.Registerer
	// StatusCodeLabel is the name that will be set to the status code label, by default is `code`.
	StatusCodeLabel string
	// MethodLabel is the name that will be set to the method label, by default is `method`.
	MethodLabel string
	// ServiceLabel is the name that will be set to the service label, by default is `service`.
	ServiceLabel string
	// HandlerIDLabel is the name that will be set to the handler ID label, by default is `handler`.
	HandlerIDLabel string
}

func (procfg *PrometheusConfig) defaults() {
	if procfg.HandlerIDLabel == "" {
		procfg.HandlerIDLabel = "handler"
	}
	if procfg.ServiceLabel == "" {
		procfg.ServiceLabel = "service"
	}
	if procfg.MethodLabel == "" {
		procfg.MethodLabel = "method"
	}
	if procfg.StatusCodeLabel == "" {
		procfg.StatusCodeLabel = "code"
	}
	if procfg.Registry == nil {
		procfg.Registry = prometheus.DefaultRegisterer
	}
	if len(procfg.DurationBuckets) == 0 {
		procfg.DurationBuckets = prometheus.DefBuckets
	}
}

// NewPrometheus - create new prometheus recorder
func NewPrometheus(cfg PrometheusConfig) Recorder {
	cfg.defaults()
	labels := []string{cfg.ServiceLabel, cfg.HandlerIDLabel, cfg.MethodLabel, cfg.StatusCodeLabel}
	r := &prometheusRecorder{
		requestLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Prefix,
			Subsystem: "http",
			Name:      "request_duration_second",
			Help:      "The latency of http request",
			Buckets:   cfg.DurationBuckets,
		}, labels),
		requestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Prefix,
			Subsystem: "http",
			Name:      "request_count",
			Help:      "Number of requests received",
		}, labels),
		requestCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: cfg.Prefix,
			Subsystem: "http",
			Name:      "request_current",
			Help:      "Current number of requests to a specific service",
		}, []string{cfg.ServiceLabel, cfg.HandlerIDLabel, cfg.MethodLabel}),
	}
	cfg.Registry.MustRegister(r.requestLatency, r.requestCounter, r.requestCurrent)
	return r
}

func (p *prometheusRecorder) ObserveHTTPRequestDuration(_ context.Context, props HTTPReqProperties, duration time.Duration) {
	p.requestLatency.WithLabelValues(props.ServiceName, props.Handler, props.Method, props.Code).Observe(duration.Seconds())
}

func (p *prometheusRecorder) ObserveHTTPRequestCount(_ context.Context, props HTTPReqProperties) {
	p.requestCounter.WithLabelValues(props.ServiceName, props.Handler, props.Method, props.Code).Add(1)
}

func (p *prometheusRecorder) IncCurrentRequest(_ context.Context, props HTTPReqProperties) {
	p.requestCurrent.WithLabelValues(props.ServiceName, props.Handler, props.Method).Inc()
}

func (p *prometheusRecorder) DescCurrentRequest(_ context.Context, props HTTPReqProperties) {
	p.requestCurrent.WithLabelValues(props.ServiceName, props.Handler, props.Method).Dec()
}
