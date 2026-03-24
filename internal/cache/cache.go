package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "calorie-proxy/internal/models"

    "github.com/go-redis/redis/v8"
)

type RedisCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewRedisCache(addr, password string, db int, ttl, connectTimeout time.Duration) (*RedisCache, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to redis: %w", err)
    }

    return &RedisCache{
        client: client,
        ttl:    ttl,
    }, nil
}

// GetFood получает продукт из кэша
func (c *RedisCache) GetFood(ctx context.Context, query string) (*models.FoodResponse, error) {
    key := fmt.Sprintf("food:%s", query)
    
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, nil // Ключ не найден
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get from cache: %w", err)
    }

    var response models.FoodResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
    }

    response.Cached = true
    return &response, nil
}

// SetFood сохраняет продукт в кэш
func (c *RedisCache) SetFood(ctx context.Context, query string, response *models.FoodResponse) error {
    key := fmt.Sprintf("food:%s", query)
    
    data, err := json.Marshal(response)
    if err != nil {
        return fmt.Errorf("failed to marshal response: %w", err)
    }

    return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Get получает любой объект из кэша (обобщенная версия)
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil
    }
    if err != nil {
        return fmt.Errorf("failed to get from cache: %w", err)
    }

    if err := json.Unmarshal(data, dest); err != nil {
        return fmt.Errorf("failed to unmarshal: %w", err)
    }

    return nil
}

// Set сохраняет любой объект в кэш
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("failed to marshal: %w", err)
    }

    return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Delete удаляет ключ из кэша
func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

// Close закрывает соединение с Redis
func (c *RedisCache) Close() error {
    return c.client.Close()
}