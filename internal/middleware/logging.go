package middleware

import (
    "log"
    "net/http"
    "time"
)

// responseWriter обертка для захвата статус кода
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware для логирования запросов
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Создаем обертку для захвата статус кода
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Вызываем следующий обработчик
        next.ServeHTTP(wrapped, r)
        
        // Логируем запрос
        log.Printf(
            "[%s] %s %s %s %d %v",
            r.Method,
            r.URL.Path,
            r.URL.RawQuery,
            r.RemoteAddr,
            wrapped.statusCode,
            time.Since(start),
        )
    })
}