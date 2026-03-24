// internal/clients/factory.go
package clients

import (
    "fmt"

    "calorie-proxy/internal/config"
)

type ProviderType string

const (
    OpenFoodFacts ProviderType = "openfoodfacts"
)

// NewProvider создает провайдера данных по типу (с опциональным конфигом)
func NewProvider(providerType ProviderType, cfg *config.Config) (FoodDataProvider, error) {
    if cfg == nil {
        cfg = &config.Config{} // дефолты через Load() не вызываем, используем пустой
    }
    switch providerType {
    case OpenFoodFacts:
        return NewOpenFoodFactsClientWithOptions(cfg.RequestTimeout, cfg.OpenFoodFactsBaseURL, cfg.OpenFoodFactsUserAgent), nil
    default:
        return nil, fmt.Errorf("unknown provider type: %s", providerType)
    }
}

// NewDefaultProvider создает провайдера по умолчанию (без конфига)
func NewDefaultProvider() FoodDataProvider {
    return NewOpenFoodFactsClient()
}

// NewDefaultProviderFromConfig создает провайдера по умолчанию с настройками из env
func NewDefaultProviderFromConfig(cfg *config.Config) FoodDataProvider {
    if cfg == nil {
        return NewDefaultProvider()
    }
    p, _ := NewProvider(OpenFoodFacts, cfg)
    if p == nil {
        return NewDefaultProvider()
    }
    return p
}