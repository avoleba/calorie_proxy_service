package clients

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strconv"
    "time"
	"math"
	"log"

    "calorie-proxy/internal/models"
)

type OpenFoodFactsClient struct {
    httpClient *http.Client
    baseURL    string
    userAgent  string
}

const (
    defaultOFFTimeout  = 10 * time.Second
    defaultOFFBaseURL  = "https://world.openfoodfacts.org"
    defaultOFFUserAgent = "CalorieProxy/1.0"
)

// NewOpenFoodFactsClient создаёт клиент с дефолтными настройками
func NewOpenFoodFactsClient() *OpenFoodFactsClient {
    return NewOpenFoodFactsClientWithOptions(defaultOFFTimeout, defaultOFFBaseURL, defaultOFFUserAgent)
}

// NewOpenFoodFactsClientWithOptions создаёт клиент с заданными timeout, baseURL, userAgent (пустые значения заменяются дефолтами)
func NewOpenFoodFactsClientWithOptions(timeout time.Duration, baseURL, userAgent string) *OpenFoodFactsClient {
    if timeout <= 0 {
        timeout = defaultOFFTimeout
    }
    if baseURL == "" {
        baseURL = defaultOFFBaseURL
    }
    if userAgent == "" {
        userAgent = defaultOFFUserAgent
    }
    return &OpenFoodFactsClient{
        httpClient: &http.Client{Timeout: timeout},
        baseURL:    baseURL,
        userAgent:  userAgent,
    }
}

func (c *OpenFoodFactsClient) Name() string {
    return "openfoodfacts"
}

func (c *OpenFoodFactsClient) SearchFoods(ctx context.Context, query string, page, pageSize int) ([]models.FoodItem, int, error) {
    if pageSize > 50 {
        pageSize = 50
    }

    encodedQuery := url.QueryEscape(query)
    url := fmt.Sprintf("%s/cgi/search.pl?search_terms=%s&page=%d&page_size=%d&json=1",
        c.baseURL, encodedQuery, page, pageSize)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("User-Agent", c.userAgent)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, 0, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var rawResult map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&rawResult); err != nil {
        return nil, 0, fmt.Errorf("failed to decode response: %w", err)
    }

    var totalCount int
    if countVal, ok := rawResult["count"]; ok {
        switch v := countVal.(type) {
        case float64:
            totalCount = int(v)
        case string:
            totalCount, _ = strconv.Atoi(v)
        default:
            totalCount = 0
        }
    }

    var items []models.FoodItem
    if products, ok := rawResult["products"].([]interface{}); ok {
        for _, p := range products {
            product, ok := p.(map[string]interface{})
            if !ok {
                continue
            }

            name, _ := product["product_name"].(string)
            if name == "" {
                continue
            }

            id, _ := product["_id"].(string)
            
            brand, _ := product["brands"].(string)
            
            categories, _ := product["categories"].(string)
            
            imageURL, _ := product["image_url"].(string)

            nutrition := models.NutritionInfo{}
            if nutriments, ok := product["nutriments"].(map[string]interface{}); ok {
                nutrition.Calories = parseFloat(nutriments["energy-kcal_100g"])
				nutrition.Protein = parseFloat(nutriments["proteins_100g"])
				nutrition.Fat = parseFloat(nutriments["fat_100g"])
				nutrition.Carbohydrates = parseFloat(nutriments["carbohydrates_100g"])
				nutrition.Fiber = parseFloat(nutriments["fiber_100g"])
				nutrition.Sugar = parseFloat(nutriments["sugars_100g"])
				nutrition.Sodium = parseFloat(nutriments["sodium_100g"])
            }

            item := models.FoodItem{
                ID:       id,
                Name:     name,
                Brand:    brand,
                Category: categories,
                Nutrition: nutrition,
                ImageURL: imageURL,
                Source:   c.Name(),
            }
            items = append(items, item)
        }
    }

    return items, totalCount, nil
}

func (c *OpenFoodFactsClient) GetFoodByBarcode(ctx context.Context, barcode string) (*models.FoodItem, error) {
    url := fmt.Sprintf("%s/api/v0/product/%s.json", c.baseURL, barcode)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("User-Agent", c.userAgent)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var rawResult map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&rawResult); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    status, _ := rawResult["status"].(float64)
    if status != 1 {
        return nil, fmt.Errorf("product not found")
    }

    product, ok := rawResult["product"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid product data")
    }

    name, _ := product["product_name"].(string)
    if name == "" {
        return nil, fmt.Errorf("product has no name")
    }

    id, _ := product["_id"].(string)
    brand, _ := product["brands"].(string)
    categories, _ := product["categories"].(string)
    imageURL, _ := product["image_url"].(string)

    nutrition := models.NutritionInfo{}
    if nutriments, ok := product["nutriments"].(map[string]interface{}); ok {
        nutrition.Calories = parseFloat(nutriments["energy-kcal_100g"])
		nutrition.Protein = parseFloat(nutriments["proteins_100g"])
		nutrition.Fat = parseFloat(nutriments["fat_100g"])
		nutrition.Carbohydrates = parseFloat(nutriments["carbohydrates_100g"])
		nutrition.Fiber = parseFloat(nutriments["fiber_100g"])
		nutrition.Sugar = parseFloat(nutriments["sugars_100g"])
		nutrition.Sodium = parseFloat(nutriments["sodium_100g"])
    }

    return &models.FoodItem{
        ID:       id,
        Name:     name,
        Brand:    brand,
        Category: categories,
        Nutrition: nutrition,
        ImageURL: imageURL,
        Source:   c.Name(),
    }, nil
}

func parseFloat(value interface{}) float64 {
    if value == nil {
        return 0
    }
    
    switch v := value.(type) {
    case float64:
        return math.Round(v*100) / 100
        
    case float32:
        return math.Round(float64(v)*100) / 100
        
    case string:
        if v == "" {
            return 0
        }
        if parsed, err := strconv.ParseFloat(v, 64); err == nil {
            return math.Round(parsed*100) / 100
        }
        log.Printf("Failed to parse float from string: %s", v)
        return 0
        
    case int:
        return float64(v)
        
    case int64:
        return float64(v)
        
    case int32:
        return float64(v)
        
    case bool:
        if v {
            return 1
        }
        return 0
        
    default:
        log.Printf("Unexpected type for float value: %T, value: %v", v, v)
        return 0
    }
}