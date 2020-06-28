package redis_test

import (
	"github.com/barcodepro/go-instrumenting/metrics"
	redismetrics "github.com/barcodepro/go-instrumenting/metrics/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRedisRecorder(t *testing.T) {
	testCases := []struct {
		name          string
		config        redismetrics.Config
		recordMetrics func(r metrics.RedisRecorder)
		expMetrics    []string
	}{
		{
			name:   "Default configuration should measure with the default metric style.",
			config: redismetrics.Config{},
			recordMetrics: func(r metrics.RedisRecorder) {
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "SET", Code: "ok"}, 10*time.Millisecond)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "SET", Code: "ok"}, 25*time.Millisecond)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "SET", Code: "ok"}, 80*time.Millisecond)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "GET", Code: "err"}, 10*time.Millisecond)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "GET", Code: "err"}, 30*time.Millisecond)
			},
			expMetrics: []string{
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.005"} 0`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.01"} 1`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.025"} 2`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.05"} 2`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.1"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.25"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="0.5"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="1"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="2.5"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="5"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="10"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="+Inf"} 3`,
				`app_redis_request_duration_seconds_sum{command="SET",keyspace="example",service="test-app",status="ok"} 0.115`,
				`app_redis_request_duration_seconds_count{command="SET",keyspace="example",service="test-app",status="ok"} 3`,

				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.005"} 0`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.01"} 1`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.025"} 1`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.05"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.1"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.25"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="0.5"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="1"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="2.5"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="5"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="10"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="+Inf"} 2`,
				`app_redis_request_duration_seconds_sum{command="GET",keyspace="example",service="test-app",status="err"} 0.04`,
				`app_redis_request_duration_seconds_count{command="GET",keyspace="example",service="test-app",status="err"} 2`,

				`app_redis_requests_total{command="SET",keyspace="example",service="test-app",status="ok"} 3`,
				`app_redis_requests_total{command="GET",keyspace="example",service="test-app",status="err"} 2`,
			},
		},
		{
			name: "Default configuration should measure with the default metric style.",
			config: redismetrics.Config{
				DurationBuckets: []float64{1, 2, 10, 20, 50, 200, 500, 1000, 2000, 5000, 10000},
			},
			recordMetrics: func(r metrics.RedisRecorder) {
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "SET", Code: "ok"}, 10*time.Second)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "SET", Code: "ok"}, 30*time.Second)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "SET", Code: "ok"}, 280*time.Second)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "GET", Code: "err"}, 10*time.Second)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "GET", Code: "err"}, 30*time.Second)
				r.Collect(metrics.RedisReqProperties{Keyspace: "example", Command: "GET", Code: "err"}, 500*time.Second)
			},
			expMetrics: []string{
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="1"} 0`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="2"} 0`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="10"} 1`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="20"} 1`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="50"} 2`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="200"} 2`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="500"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="1000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="2000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="5000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="10000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="SET",keyspace="example",service="test-app",status="ok",le="+Inf"} 3`,
				`app_redis_request_duration_seconds_sum{command="SET",keyspace="example",service="test-app",status="ok"} 320`,
				`app_redis_request_duration_seconds_count{command="SET",keyspace="example",service="test-app",status="ok"} 3`,

				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="1"} 0`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="2"} 0`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="10"} 1`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="20"} 1`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="50"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="200"} 2`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="500"} 3`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="1000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="2000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="5000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="10000"} 3`,
				`app_redis_request_duration_seconds_bucket{command="GET",keyspace="example",service="test-app",status="err",le="+Inf"} 3`,
				`app_redis_request_duration_seconds_sum{command="GET",keyspace="example",service="test-app",status="err"} 540`,
				`app_redis_request_duration_seconds_count{command="GET",keyspace="example",service="test-app",status="err"} 3`,

				`app_redis_requests_total{command="GET",keyspace="example",service="test-app",status="err"} 3`,
				`app_redis_requests_total{command="SET",keyspace="example",service="test-app",status="ok"} 3`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			metricRecorder := redismetrics.NewRedisRecorder("test-app", tc.config)
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
