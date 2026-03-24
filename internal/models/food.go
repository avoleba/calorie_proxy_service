package models

import "time"

// FoodRequest - запрос на поиск продукта
type FoodRequest struct {
    Query    string `json:"query" form:"query" binding:"required"`
    Lang     string `json:"lang" form:"lang" default:"en"` // язык для результатов
    Page     int    `json:"page" form:"page" default:"1"`
    PageSize int    `json:"page_size" form:"page_size" default:"10"`
}

// NutritionInfo - информация о пищевой ценности (на 100г)
type NutritionInfo struct {
    Calories      float64 `json:"calories"`       // ккал
    Protein       float64 `json:"protein"`        // белки (г)
    Fat           float64 `json:"fat"`            // жиры (г)
    Carbohydrates float64 `json:"carbohydrates"`  // углеводы (г)
    Fiber         float64 `json:"fiber,omitempty"` // клетчатка (г)
    Sugar         float64 `json:"sugar,omitempty"` // сахар (г)
    Sodium        float64 `json:"sodium,omitempty"` // натрий (мг)
}

// FoodItem - информация о продукте
type FoodItem struct {
    ID          string       `json:"id"`                    // уникальный ID продукта
    Name        string       `json:"name"`                   // название продукта
    Brand       string       `json:"brand,omitempty"`        // бренд (для брендированных)
    Category    string       `json:"category,omitempty"`     // категория
    Description string       `json:"description,omitempty"`  // описание
    Nutrition   NutritionInfo `json:"nutrition"`              // пищевая ценность на 100г
    ServingSize string       `json:"serving_size,omitempty"` // размер порции
    ImageURL    string       `json:"image_url,omitempty"`     // URL изображения
    Source      string       `json:"source"`                  // источник данных (usda, openfoodfacts, edamam)
}

// FoodResponse - ответ на поиск продуктов
type FoodResponse struct {
    Items     []FoodItem `json:"items"`
    Total     int        `json:"total"`
    Page      int        `json:"page"`
    PageSize  int        `json:"page_size"`
    Cached    bool       `json:"cached,omitempty"`
    Timestamp time.Time  `json:"timestamp"`
}

// ErrorResponse - ответ с ошибкой
type ErrorResponse struct {
    Error   string `json:"error"`
    Status  int    `json:"status"`
    Message string `json:"message,omitempty"`
}

// NutritionAnalysisRequest - запрос на анализ рецепта/блюда
type NutritionAnalysisRequest struct {
    Ingredients []string `json:"ingredients" binding:"required"` // список ингредиентов
    Title       string   `json:"title,omitempty"`                // название блюда
}

// NutritionAnalysisResponse - результат анализа блюда
type NutritionAnalysisResponse struct {
    Title       string       `json:"title"`
    TotalWeight float64      `json:"total_weight"`        // общий вес (г)
    Yield       int          `json:"yield"`               // количество порций
    Calories    float64      `json:"calories"`            // всего калорий
    TotalNutrients NutritionInfo `json:"total_nutrients"` // всего БЖУ
    PerServing  NutritionInfo `json:"per_serving"`        // на порцию
    Ingredients []FoodItem    `json:"ingredients"`        // использованные продукты
    Cached      bool          `json:"cached,omitempty"`
    Timestamp   time.Time     `json:"timestamp"`
}