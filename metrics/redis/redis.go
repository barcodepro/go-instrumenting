package redis

import (
	"context"
	"github.com/barcodepro/go-instrumenting/metrics"
	"github.com/go-redis/redis/v7"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strings"
	"time"
)

const (
	labelApp      = "application"
	labelStatus   = "status"
	labelCommand  = "command"
	labelKeyspace = "keyspace"

	keyRequestStart key = iota
)

type key int

type Config struct {
	// DurationBuckets are the buckets used by Prometheus for the HTTP request duration metrics,
	// by default uses Prometheus default buckets (from 5ms to 10s).
	DurationBuckets []float64
}

func (c *Config) defaults() {
	if len(c.DurationBuckets) == 0 {
		c.DurationBuckets = prometheus.DefBuckets
	}
}

type recorder struct {
	Registry                        prometheus.Registerer
	RedisRequestsTotal              *prometheus.CounterVec
	RedisRequestsDurationsHistogram *prometheus.HistogramVec
}

func NewRedisRecorder(appName string, config Config) metrics.RedisRecorder {
	config.defaults()

	r := &recorder{
		RedisRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "app",
			Subsystem:   "redis",
			Name:        "requests_total",
			Help:        "The total number of processed requests.",
			ConstLabels: map[string]string{labelApp: appName},
		}, []string{labelCommand, labelKeyspace, labelStatus}),

		RedisRequestsDurationsHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "app",
			Subsystem:   "redis",
			Name:        "request_duration_seconds",
			Help:        "The latency of the Redis requests.",
			Buckets:     config.DurationBuckets,
			ConstLabels: map[string]string{labelApp: appName},
		}, []string{labelCommand, labelKeyspace, labelStatus}),
	}

	r.Registry = prometheus.DefaultRegisterer

	r.Registry.MustRegister(
		r.RedisRequestsTotal,
		r.RedisRequestsDurationsHistogram,
	)

	return r
}

// Collect updates metrics using passed properties
func (r recorder) Collect(props metrics.RedisReqProperties, duration time.Duration) {
	var (
		code    = props.Code
		command = props.Command
		space   = props.Keyspace
	)

	r.RedisRequestsTotal.WithLabelValues(command, space, code).Inc()
	r.RedisRequestsDurationsHistogram.WithLabelValues(command, space, code).Observe(duration.Seconds())
}

// Unregister ...
func (r recorder) Unregister() {
	r.Registry.Unregister(r.RedisRequestsTotal)
	r.Registry.Unregister(r.RedisRequestsDurationsHistogram)
}

func (r recorder) NewCollectHook() redis.Hook {
	hook := &CollectHook{
		recorder: r,
	}
	return redis.Hook(hook)
}

// CollectHook is an implementation of redis.Hook interface
type CollectHook struct {
	recorder
}

func (h *CollectHook) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, keyRequestStart, time.Now())
	return ctx, nil
}

func (h *CollectHook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	// no-op method
	return ctx, nil
}

func (h *CollectHook) AfterProcessPipeline(_ context.Context, _ []redis.Cmder) error {
	// no-op method
	return nil
}

func (h *CollectHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	var props = metrics.RedisReqProperties{
		Command: cmd.Name(),
	}
	if cmd.Err() != nil {
		props.Code = "err"
	} else {
		props.Code = "ok"
	}

	// If number of passed arguments greater than than 1 it means we can extract key of GET/SET/DEL/etc commands.
	// Regexp adjusted to the following format of keys - 'app-name/keyspace/...', hence using the regexp below, the app-name
	// will be omitted (app-name is already passed as a part of metric) and rest of key will be extracted.
	if len(cmd.Args()) >= 2 {
		argsStr := strings.Split(cmd.String(), " ")
		re := regexp.MustCompile(`(/[a-z-]{1,})+`)
		props.Keyspace = re.FindString(argsStr[1])
	}

	// Extract request start time from context
	start := ctx.Value(keyRequestStart).(time.Time)

	h.recorder.Collect(props, time.Since(start))

	return nil
}
