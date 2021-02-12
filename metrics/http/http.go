package http

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaponry/go-instrumenting/metrics"
	"time"
)

const (
	labelApp    = "application"
	labelPath   = "path"
	labelMethod = "method"
	labelStatus = "status"
)

type Config struct {
	// DurationBuckets are the buckets used by Prometheus for the HTTP request duration metrics,
	// by default uses Prometheus default buckets (from 5ms to 10s).
	DurationBuckets []float64
	// SizeBuckets are the buckets used by Prometheus for the HTTP response size metrics,
	// by default uses a exponential buckets from 100B to 1GB.
	SizeBuckets []float64
}

func (c *Config) defaults() {
	if len(c.DurationBuckets) == 0 {
		c.DurationBuckets = prometheus.DefBuckets
	}

	if len(c.SizeBuckets) == 0 {
		c.SizeBuckets = prometheus.ExponentialBuckets(100, 10, 8)
	}
}

type recorder struct {
	Registry                       prometheus.Registerer
	HttpRequestsTotal              *prometheus.CounterVec
	HttpRequestsDurationsHistogram *prometheus.HistogramVec
	HttpResponseSizeHistogram      *prometheus.HistogramVec
}

func NewHttpRecorder(appName string, config Config) metrics.HttpRecorder {
	config.defaults()

	r := &recorder{
		HttpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "app",
			Subsystem:   "http",
			Name:        "requests_total",
			Help:        "The total number of processed requests.",
			ConstLabels: map[string]string{labelApp: appName},
		}, []string{labelPath, labelMethod, labelStatus}),

		HttpRequestsDurationsHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "app",
			Subsystem:   "http",
			Name:        "request_duration_seconds",
			Help:        "The latency of the HTTP requests.",
			Buckets:     config.DurationBuckets,
			ConstLabels: map[string]string{labelApp: appName},
		}, []string{labelPath, labelMethod, labelStatus}),

		HttpResponseSizeHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "app",
			Subsystem:   "http",
			Name:        "response_size_bytes",
			Help:        "The size of the HTTP responses.",
			Buckets:     config.SizeBuckets,
			ConstLabels: map[string]string{labelApp: appName},
		}, []string{labelPath, labelMethod, labelStatus}),
	}

	r.Registry = prometheus.DefaultRegisterer

	r.Registry.MustRegister(
		r.HttpRequestsTotal,
		r.HttpRequestsDurationsHistogram,
		r.HttpResponseSizeHistogram,
	)

	return r
}

// Collect updates metrics using passed properties
func (r recorder) Collect(props metrics.HTTPReqProperties, duration time.Duration, bytesWritten int) {
	r.HttpRequestsTotal.WithLabelValues(props.Path, props.Method, props.Code).Inc()
	r.HttpRequestsDurationsHistogram.WithLabelValues(props.Path, props.Method, props.Code).Observe(duration.Seconds())
	r.HttpResponseSizeHistogram.WithLabelValues(props.Path, props.Method, props.Code).Observe(float64(bytesWritten))
}

// Unregister ...
func (r recorder) Unregister() {
	r.Registry.Unregister(r.HttpRequestsTotal)
	r.Registry.Unregister(r.HttpRequestsDurationsHistogram)
	r.Registry.Unregister(r.HttpResponseSizeHistogram)
}
