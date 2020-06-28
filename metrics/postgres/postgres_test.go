package postgres_test

import (
	"github.com/barcodepro/go-instrumenting/metrics"
	postgresmetrics "github.com/barcodepro/go-instrumenting/metrics/postgres"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPostgresRecorder(t *testing.T) {
	testCases := []struct {
		name          string
		recordMetrics func(r metrics.PostgresRecorder)
		expMetrics    []string
	}{
		{
			name: "Default configuration should measure with the default metric style.",
			recordMetrics: func(r metrics.PostgresRecorder) {
				r.Collect()
			},
			expMetrics: []string{
				`app_postgres_xacts_total{service="test-app"} 1`,
			},
		},
		{
			name: "Default configuration should measure with the default metric style.",
			recordMetrics: func(r metrics.PostgresRecorder) {
				r.Collect()
			},
			expMetrics: []string{
				`app_postgres_xacts_total{service="test-app"} 1`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metricRecorder := postgresmetrics.NewPostgresRecorder("test-app")
			tc.recordMetrics(metricRecorder)

			// Get the metrics handler and serve.
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}).ServeHTTP(rec, req)

			resp := rec.Result()

			// Check all metrics are present.
			if assert.Equal(t, http.StatusOK, resp.StatusCode) {
				body, _ := ioutil.ReadAll(resp.Body)
				for _, expMetric := range tc.expMetrics {
					assert.Contains(t, string(body), expMetric, "metric not present on the result")
				}
			}
			metricRecorder.Unregister()
		})
	}
}
