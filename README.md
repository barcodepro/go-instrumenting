# go-instrumenting
Libraries for instrumenting Go applications.

- HTTP metrics based on widely used middleware approach.
- Redis metrics based on `Hook` interface of [go-redis/redis](https://github.com/go-redis/redis).
- Postgres metrics based on `AfterRelease` function of [jackc/pgx](https://github.com/jackc/pgx) pools.

[Examples](#usage-examples):
- [http](#instrumenting-http)
- [postgres and redis](#instrumenting-postgres-and-redis)

---

#### HTTP metrics
HTTP metrics use middleware for processing requests and collecting data. Any convenient middlewares can be used, see example below.
```
# COUNTER app_http_requests_total The total number of processed requests.
# HISTOGRAM app_http_request_duration_seconds The latency of the HTTP requests.
# HISTOGRAM app_http_response_size_bytes The size of the HTTP responses.
```

#### Redis metrics
Redis metrics are collected using `Hook` interface provided by [go-redis/redis](https://github.com/go-redis/redis).

```
# COUNTER app_redis_requests_total The total number of processed requests.
# HISTOGRAM app_redis_request_duration_seconds The latency of the Redis requests.
```

#### Postgres metrics
Postgres metrics are collected using `AfterRelease` function provided by [jackc/pgx](https://github.com/jackc/pgx) pools. Cuurently this is the poorest way to collect metrics. Hope things getting better [later](https://github.com/jackc/pgx/issues/782).
```
# COUNTER app_postgres_xacts_total The total number of processed transactions.
```

#### Usage examples

##### Instrumenting HTTP:
Declare metric recorder.
```
// Main server struct
type server struct {
    httpserver *http.Server
    metrics    metrics.HttpRecorder	
    ...
}

// Optional response writer used for hijacking http.ResponseWriter in middlewares
type responseWriter struct {
	http.ResponseWriter
	code         int
	bytesWritten int
}
```
Create new HTTP recorder (it register metrics under the hood):
```
s := &server{
    metrics = httpmetrics.NewHttpRecorder("myService", httpmetrics.Config{})
}
```
Create middleware using the way you prefer. In the middleware extract properties from request, pass them into deferred Collect method executed after request has been handled. 
```
func (s *server) instrumentRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rw := &responseWriter{w, http.StatusOK, 0}

        start := time.Now()
        defer func() {
            props := metrics.HTTPReqProperties{
                Path:   r.RequestURI,
                Method: r.Method,
                Code:   rw.code,
            }
            s.metrics.Collect(props, time.Since(start), rw.bytesWritten)
        }()

        next.ServeHTTP(rw, r)
    })
}
```

##### Instrumenting Postgres and Redis:
Create Metric wrapper with metrics recorders, put created wrapper into main store struct.
```
// Metrics struct wraps all store-related metrics recorders
type Metrics struct {
	RedisMetrics    metrics.RedisRecorder
	PostgresMetrics metrics.PostgresRecorder
}

// Main store struct
type Store struct {
	PgDB    *pgxpool.Pool
	RedisDB *redis.Client
	Metrics Metrics
}
```
Create recorders in function where store is created
```
func NewStore(ctx context.Context, c *Config) (*Store, error) {
	var s = new(Store)

	s.Metrics.RedisMetrics = redismetrics.NewRedisRecorder("MyServive", redismetrics.Config{})
	s.Metrics.PostgresMetrics = postgresmetrics.NewPostgresRecorder("MyService")

	pgdbStore, err := NewPostgresStore(c.PostgresURL, s.Metrics.PostgresMetrics)
	if err != nil {
        return nil, err    
    }
    s.PgDB = pgdbStore

    redisStore, err := NewRedisStore(c.RedisURL, s.Metrics.RedisMetrics)
	if err != nil {
        return nil, err
    }
    s.RedisDB = redisStore

    // other store creation logic
 
	return s, nil
}
```
For Postgres, assign AfterRelease function to AfterRelease of pgxpool.Config.
```
func NewPostgresStore(postgresURL string, metrics metrics.PostgresRecorder) (*pgxpool.Pool, error) {
	pgConfig, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		return nil, err
	}

	pgConfig.AfterRelease = metrics.AfterReleaseHook
	
	return pgxpool.ConnectConfig(context.Background(), pgConfig)
}
```
For Redis, add Hook to the client.
```
func NewRedisStore(redisURL string, metrics metrics.RedisRecorder) (*redis.Client, error) {
	redisConfig, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(redisConfig)
	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}

	// Initialize metrics subsystem.
	client.AddHook(metrics.NewCollectHook())

	return client, nil
}
```
