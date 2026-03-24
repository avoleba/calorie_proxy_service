package proxy

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
	"strconv"

    "calorie-proxy/internal/cache"
    "calorie-proxy/internal/clients"
    "calorie-proxy/internal/config"
    "calorie-proxy/internal/middleware"
    "calorie-proxy/internal/models"
)

type ProxyHandler struct {
    config   *config.Config
    cache    *cache.RedisCache
    provider clients.FoodDataProvider  // Добавляем провайдера
}

func NewProxyHandler(cfg *config.Config, cache *cache.RedisCache) *ProxyHandler {
    provider := clients.NewDefaultProviderFromConfig(cfg)
    return &ProxyHandler{
        config:   cfg,
        cache:    cache,
        provider: provider,
    }
}

// SearchFoodHandler - поиск продуктов через реальное API
func (h *ProxyHandler) SearchFoodHandler(w http.ResponseWriter, r *http.Request) {    
    query := r.URL.Query().Get("q")
    if query == "" {
        h.sendError(w, http.StatusBadRequest, "missing query parameter 'q'")
        return
    }

    page := 1
    pageSize := 10

    if pageStr := r.URL.Query().Get("page"); pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
        }
    }
    
    if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
        if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
            if ps <= 100 {
                pageSize = ps
            }
        }
    }

    ctx := r.Context()
    cacheKey := fmt.Sprintf("search:%s:%d:%d", query, page, pageSize)
    
    var cachedResponse models.FoodResponse
    err := h.cache.Get(ctx, cacheKey, &cachedResponse)
    if err == nil && cachedResponse.Items != nil {
        middleware.RecordCacheHit()
        cachedResponse.Cached = true
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Cache", "HIT")
        json.NewEncoder(w).Encode(cachedResponse)
        return
    }
    middleware.RecordCacheMiss()

    items, total, err := h.provider.SearchFoods(ctx, query, page, pageSize)
    if err != nil {
        h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("API error: %v", err))
        return
    }

    response := &models.FoodResponse{
        Items:     items,
        Total:     total,
        Page:      page,
        PageSize:  pageSize,
        Timestamp: time.Now(),
    }

    go func() {
        h.cache.Set(ctx, cacheKey, response)
    }()

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Cache", "MISS")
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        h.sendError(w, http.StatusInternalServerError, "failed to encode response")
    }
}

// BarcodeHandler - поиск по штрих-коду через реальное API
func (h *ProxyHandler) BarcodeHandler(w http.ResponseWriter, r *http.Request) {    
    barcode := r.URL.Query().Get("barcode")
    if barcode == "" {
        h.sendError(w, http.StatusBadRequest, "missing barcode parameter")
        return
    }

    ctx := r.Context()
    cacheKey := fmt.Sprintf("barcode:%s", barcode)
    
    // Проверяем кэш
    var cachedItem models.FoodItem
    err := h.cache.Get(ctx, cacheKey, &cachedItem)
    if err == nil && cachedItem.ID != "" {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Cache", "HIT")
        json.NewEncoder(w).Encode(cachedItem)
        return
    }

    // Ищем через API
    item, err := h.provider.GetFoodByBarcode(ctx, barcode)
    if err != nil {
        h.sendError(w, http.StatusNotFound, fmt.Sprintf("Product not found: %v", err))
        return
    }

    // Сохраняем в кэш
    go func() {
        h.cache.Set(ctx, cacheKey, item)
    }()

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Cache", "MISS")
    json.NewEncoder(w).Encode(item)
}

// ... остальные методы (HealthHandler, sendError, sendRateLimitResponse) без изменений

// HealthHandler - проверка здоровья
func (h *ProxyHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "ok",
        "time":   time.Now().Format(time.RFC3339),
    })
}

// Вспомогательные функции
func (h *ProxyHandler) sendError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(models.ErrorResponse{
        Error:   http.StatusText(status),
        Status:  status,
        Message: message,
    })
}

func (h *ProxyHandler) sendRateLimitResponse(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusTooManyRequests)
    
    json.NewEncoder(w).Encode(models.ErrorResponse{
        Error:   "rate_limit_exceeded",
        Status:  http.StatusTooManyRequests,
        Message: "Too many requests. Please try again later.",
    })
}