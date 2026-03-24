// cmd/proxy/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"calorie-proxy/internal/auth"
	"calorie-proxy/internal/cache"
	"calorie-proxy/internal/cart"
	"calorie-proxy/internal/config"
	"calorie-proxy/internal/middleware"
	"calorie-proxy/internal/proxy"
	"calorie-proxy/internal/store"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	db, err := store.OpenDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	st := store.NewStore(db)

	redisCache, err := cache.NewRedisCache(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
		cfg.CacheTTL,
		cfg.RedisConnectTimeout,
	)
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
	defer redisCache.Close()

	proxyHandler := proxy.NewProxyHandler(cfg, redisCache)
	authHandler := auth.NewHandler(st, cfg.JWTSecret)
	cartHandler := cart.NewHandler(st)

	requireAuth := auth.RequireAuth(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/foods/search", requireAuth(http.HandlerFunc(proxyHandler.SearchFoodHandler)))
	mux.Handle("/api/v1/foods/barcode", requireAuth(http.HandlerFunc(proxyHandler.BarcodeHandler)))
	mux.HandleFunc("/health", proxyHandler.HealthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/api/v1/auth/register", authHandler.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", authHandler.LoginHandler)

	cartWithAuth := requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			cartHandler.AddHandler(w, r)
		case http.MethodGet:
			cartHandler.ListHandler(w, r)
		case http.MethodPatch, http.MethodPut:
			cartHandler.UpdateHandler(w, r)
		case http.MethodDelete:
			cartHandler.DeleteHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cart", cartWithAuth)

	handler := middleware.Cors(cfg)(mux)
	handler = middleware.Logging(handler)
	handler = middleware.Metrics(handler)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

    go func() {
        log.Printf("Starting server on port %s", cfg.ServerPort)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}