package metrics

import (
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v4"
	"time"
)

/*
 * HTTP metrics recorder
 */

// HTTPReqProperties describes properties of HTTP requests.
type HTTPReqProperties struct {
	Path   string // URL Path of the request.
	Method string // Method of the request.
	Code   string // Response code is the request.
}

// HttpRecorder knows how to record and measure HTTP metrics.
type HttpRecorder interface {
	Collect(props HTTPReqProperties, duration time.Duration, bytesWritten int)
	Unregister()
}

/*
 * Redis metrics recorder
 */

// RedisReqProperties describes properties of Redis requests.
type RedisReqProperties struct {
	Keyspace string // Key space of the request (if key has been specified).
	Command  string // Command of the request.
	Code     string // Response code is the request.
}

// RedisRecorder knows how to record and measure Redis metrics.
type RedisRecorder interface {
	NewCollectHook() redis.Hook
	Collect(props RedisReqProperties, duration time.Duration)
	Unregister()
}

/*
 * Postgres metrics recorder
 */

// PostgresRecorder knows how to record and measure Postgres metrics.
type PostgresRecorder interface {
	AfterReleaseHook(conn *pgx.Conn) bool
	Collect()
	Unregister()
}
