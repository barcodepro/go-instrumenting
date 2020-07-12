package http_test

import (
	"github.com/barcodepro/go-instrumenting/metrics"
	httpmetrics "github.com/barcodepro/go-instrumenting/metrics/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHttpRecorder(t *testing.T) {
	testCases := []struct {
		name          string
		config        httpmetrics.Config
		recordMetrics func(r metrics.HttpRecorder)
		expMetrics    []string
	}{
		{
			name:   "Default configuration should measure with the default metric style.",
			config: httpmetrics.Config{},
			recordMetrics: func(r metrics.HttpRecorder) {
				r.Collect(metrics.HTTPReqProperties{Path: "/test", Method: http.MethodGet, Code: "200"}, 150*time.Millisecond, 500000)
				r.Collect(metrics.HTTPReqProperties{Path: "/test", Method: http.MethodPost, Code: "403"}, 10*time.Millisecond, 5000)
				r.Collect(metrics.HTTPReqProperties{Path: "/test", Method: http.MethodPost, Code: "403"}, 150*time.Millisecond, 200000)
			},
			expMetrics: []string{
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.005"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.01"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.025"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.05"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.1"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.25"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="0.5"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="1"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="2.5"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="5"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="10"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="+Inf"} 1`,
				`app_http_request_duration_seconds_sum{application="test-app",method="GET",path="/test",status="200"} 0.15`,
				`app_http_request_duration_seconds_count{application="test-app",method="GET",path="/test",status="200"} 1`,

				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.005"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.01"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.025"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.05"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.1"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.25"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="0.5"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="1"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="2.5"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="5"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="10"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="POST",path="/test",status="403",le="+Inf"} 2`,
				`app_http_request_duration_seconds_sum{application="test-app",method="POST",path="/test",status="403"} 0.16`,
				`app_http_request_duration_seconds_count{application="test-app",method="POST",path="/test",status="403"} 2`,

				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="100"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1000"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="10000"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="100000"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+06"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+07"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+08"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+09"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="+Inf"} 1`,
				`app_http_response_size_bytes_sum{application="test-app",method="GET",path="/test",status="200"} 500000`,
				`app_http_response_size_bytes_count{application="test-app",method="GET",path="/test",status="200"} 1`,

				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="100"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="1000"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="10000"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="100000"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="1e+06"} 2`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="1e+07"} 2`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="1e+08"} 2`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="1e+09"} 2`,
				`app_http_response_size_bytes_bucket{application="test-app",method="POST",path="/test",status="403",le="+Inf"} 2`,
				`app_http_response_size_bytes_sum{application="test-app",method="POST",path="/test",status="403"} 205000`,
				`app_http_response_size_bytes_count{application="test-app",method="POST",path="/test",status="403"} 2`,

				`app_http_requests_total{application="test-app",method="GET",path="/test",status="200"} 1`,
				`app_http_requests_total{application="test-app",method="POST",path="/test",status="403"} 2`,
			},
		},
		{
			name: "Custom buckets configuration should measure with the custom buckets.",
			config: httpmetrics.Config{
				DurationBuckets: []float64{1, 2, 10, 20, 50, 200, 500, 1000, 2000, 5000, 10000},
			},
			recordMetrics: func(r metrics.HttpRecorder) {
				r.Collect(metrics.HTTPReqProperties{Path: "/test", Method: http.MethodGet, Code: "200"}, 4*time.Second, 1000)
				r.Collect(metrics.HTTPReqProperties{Path: "/test", Method: http.MethodGet, Code: "200"}, 30*time.Second, 20000)
				r.Collect(metrics.HTTPReqProperties{Path: "/test", Method: http.MethodGet, Code: "200"}, 600*time.Second, 200000)
			},
			expMetrics: []string{
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="1"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="2"} 0`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="10"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="20"} 1`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="50"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="200"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="500"} 2`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="1000"} 3`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="2000"} 3`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="5000"} 3`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="10000"} 3`,
				`app_http_request_duration_seconds_bucket{application="test-app",method="GET",path="/test",status="200",le="+Inf"} 3`,
				`app_http_request_duration_seconds_sum{application="test-app",method="GET",path="/test",status="200"} 634`,
				`app_http_request_duration_seconds_count{application="test-app",method="GET",path="/test",status="200"} 3`,

				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="100"} 0`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1000"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="10000"} 1`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="100000"} 2`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+06"} 3`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+07"} 3`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+08"} 3`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="1e+09"} 3`,
				`app_http_response_size_bytes_bucket{application="test-app",method="GET",path="/test",status="200",le="+Inf"} 3`,
				`app_http_response_size_bytes_sum{application="test-app",method="GET",path="/test",status="200"} 221000`,
				`app_http_response_size_bytes_count{application="test-app",method="GET",path="/test",status="200"} 3`,

				`app_http_requests_total{application="test-app",method="GET",path="/test",status="200"} 3`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metricRecorder := httpmetrics.NewHttpRecorder("test-app", tc.config)
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
