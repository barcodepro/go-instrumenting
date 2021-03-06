package postgres

import (
	"github.com/jackc/pgx/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaponry/go-instrumenting/metrics"
)

const (
	labelApp = "application"
)

type recorder struct {
	Registry      prometheus.Registerer
	RequestsTotal *prometheus.CounterVec
}

func NewPostgresRecorder(appName string) metrics.PostgresRecorder {

	r := &recorder{
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "app",
			Subsystem:   "postgres",
			Name:        "xacts_total",
			Help:        "The total number of processed transactions.",
			ConstLabels: map[string]string{labelApp: appName},
		}, []string{}),
	}

	r.Registry = prometheus.DefaultRegisterer

	r.Registry.MustRegister(
		r.RequestsTotal,
	)

	return r
}

// Collect updates metrics using passed properties
func (r recorder) Collect() {
	r.RequestsTotal.WithLabelValues().Inc()
}

// Unregister ...
func (r recorder) Unregister() {
	r.Registry.Unregister(r.RequestsTotal)
}

func (r recorder) AfterReleaseHook(_ *pgx.Conn) bool {
	r.Collect()
	return true
}
