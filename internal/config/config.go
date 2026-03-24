package config

import (
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    ServerPort     string
    RedisAddr      string
    RedisPassword  string
    RedisDB        int
    RateDuration   time.Duration
    CacheTTL       time.Duration
    RequestTimeout time.Duration

    // Таймауты HTTP-сервера
    ServerReadTimeout  time.Duration
    ServerWriteTimeout time.Duration
    ServerIdleTimeout  time.Duration
    ShutdownTimeout    time.Duration

    // Redis
    RedisConnectTimeout time.Duration

    // CORS
    CORSAllowedOrigins string // через запятую, например "http://localhost:3000,http://127.0.0.1:3000"
    CORSMaxAge         int    // секунды

    // API Keys для разных источников
    USDAAPIKey     string
    EdamamAppID    string
    EdamamAppKey   string

    FoodProvider string

    // OpenFoodFacts (и другие клиенты при необходимости)
    OpenFoodFactsBaseURL  string
    OpenFoodFactsUserAgent string

    // Auth & DB
    JWTSecret   string
    DatabaseURL string
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    return &Config{
        ServerPort:         getEnv("SERVER_PORT", "8080"),
        RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword:      getEnv("REDIS_PASSWORD", ""),
        RedisDB:            getEnvAsInt("REDIS_DB", 0),
        RateDuration:       getEnvAsDuration("RATE_DURATION", time.Minute),
        CacheTTL:           getEnvAsDuration("CACHE_TTL", 24*time.Hour),
        RequestTimeout:     getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),
        ServerReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 5*time.Second),
        ServerWriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
        ServerIdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
        ShutdownTimeout:    getEnvAsDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
        RedisConnectTimeout: getEnvAsDuration("REDIS_CONNECT_TIMEOUT", 5*time.Second),
        CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"),
        CORSMaxAge:         getEnvAsInt("CORS_MAX_AGE", 86400),
        USDAAPIKey:         getEnv("USDA_API_KEY", ""),
        EdamamAppID:        getEnv("EDAMAM_APP_ID", ""),
        EdamamAppKey:       getEnv("EDAMAM_APP_KEY", ""),
        FoodProvider:            getEnv("FOOD_PROVIDER", "openfoodfacts"),
        OpenFoodFactsBaseURL:     getEnv("OPENFOODFACTS_BASE_URL", "https://world.openfoodfacts.org"),
        OpenFoodFactsUserAgent:   getEnv("OPENFOODFACTS_USER_AGENT", "CalorieProxy/1.0"),
        JWTSecret:                getEnv("JWT_SECRET", "dev-secret-change-in-production"),
        DatabaseURL:        getEnv("DATABASE_URL", "postgres://localhost:5432/calorie?sslmode=disable"),
    }
}

// CORSAllowedOriginsMap возвращает map разрешённых origin для быстрой проверки
func (c *Config) CORSAllowedOriginsMap() map[string]bool {
    m := make(map[string]bool)
    for _, o := range strings.Split(c.CORSAllowedOrigins, ",") {
        o = strings.TrimSpace(o)
        if o != "" {
            m[o] = true
        }
    }
    return m
}

// Хелперы для загрузки переменных окружения
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}