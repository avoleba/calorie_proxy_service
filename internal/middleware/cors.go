package middleware

import (
	"net/http"
	"strconv"

	"calorie-proxy/internal/config"
)

// Cors возвращает middleware с настройками CORS из конфига
func Cors(cfg *config.Config) func(http.Handler) http.Handler {
	allowedOrigins := cfg.CORSAllowedOriginsMap()
	defaultOrigin := ""
	for o := range allowedOrigins {
		defaultOrigin = o
		break
	}
	maxAge := strconv.Itoa(cfg.CORSMaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if defaultOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", defaultOrigin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", maxAge)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}