// internal/middleware/metrics.go
package middleware

import (
    "net/http"
    "strconv"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    cacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
    )

    cacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
    )

    rateLimitExceeded = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "rate_limit_exceeded_total",
            Help: "Total number of rate limit exceeded",
        },
    )
)

// Metrics middleware для сбора метрик
func Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Создаем обертку для захвата статус кода
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Вызываем следующий обработчик
        next.ServeHTTP(wrapped, r)
        
        // Записываем метрики
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(wrapped.statusCode)
        
        requestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
        requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

// RecordCacheHit увеличивает счетчик попаданий в кэш
func RecordCacheHit() {
    cacheHits.Inc()
}

// RecordCacheMiss увеличивает счетчик промахов кэша
func RecordCacheMiss() {
    cacheMisses.Inc()
}

// RecordRateLimitExceeded увеличивает счетчик превышений rate limit
func RecordRateLimitExceeded() {
    rateLimitExceeded.Inc()
}