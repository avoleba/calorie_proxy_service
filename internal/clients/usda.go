package clients

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "calorie-proxy/internal/models"
)

type USDAClient struct {
    httpClient *http.Client
    apiKey     string
    baseURL    string
}

func NewUSDAClient(apiKey string) *USDAClient {
    return &USDAClient{
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
        apiKey:  apiKey,
        baseURL: "https://api.nal.usda.gov/fdc/v1",
    }
}

// SearchFoods - поиск продуктов в базе USDA
func (c *USDAClient) SearchFoods(query string, page, pageSize int) ([]models.FoodItem, int, error) {
    url := fmt.Sprintf("%s/foods/search?api_key=%s", c.baseURL, c.apiKey)

    // Формируем запрос
    requestBody := map[string]interface{}{
        "query":      query,
        "pageNumber": page,
        "pageSize":   pageSize,
    }

    jsonBody, _ := json.Marshal(requestBody)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, 0, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, 0, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    // Парсим ответ USDA
    var result struct {
        Foods []struct {
            FdcID       int    `json:"fdcId"`
            Description string `json:"description"`
            BrandOwner  string `json:"brandOwner"`
            FoodNutrients []struct {
                NutrientName string  `json:"nutrientName"`
                Value        float64 `json:"value"`
                UnitName     string  `json:"unitName"`
            } `json:"foodNutrients"`
        } `json:"foods"`
        TotalCount int `json:"totalHits"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, 0, fmt.Errorf("failed to decode response: %w", err)
    }

    // Конвертируем в наши модели
    items := make([]models.FoodItem, 0, len(result.Foods))
    for _, f := range result.Foods {
        item := models.FoodItem{
            ID:    fmt.Sprintf("%d", f.FdcID),
            Name:  f.Description,
            Brand: f.BrandOwner,
            Nutrition: models.NutritionInfo{
                Calories: 0,
                Protein:  0,
                Fat:      0,
                Carbohydrates: 0,
            },
            Source: "usda",
        }

        // Извлекаем питательные вещества
        for _, nutrient := range f.FoodNutrients {
            switch nutrient.NutrientName {
            case "Energy":
                item.Nutrition.Calories = nutrient.Value
            case "Protein":
                item.Nutrition.Protein = nutrient.Value
            case "Total lipid (fat)":
                item.Nutrition.Fat = nutrient.Value
            case "Carbohydrate, by difference":
                item.Nutrition.Carbohydrates = nutrient.Value
            case "Fiber, total dietary":
                item.Nutrition.Fiber = nutrient.Value
            case "Sodium, Na":
                item.Nutrition.Sodium = nutrient.Value
            }
        }

        items = append(items, item)
    }

    return items, result.TotalCount, nil
}