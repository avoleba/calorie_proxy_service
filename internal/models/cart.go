package models

import "time"

// CartItem — позиция в корзине (в БД: продукт + граммовка)
type CartItem struct {
	ID         int64          `json:"id"`
	UserID     int64          `json:"user_id"`
	FoodID     string         `json:"food_id"`     // внешний ID продукта (из API)
	Name       string         `json:"name"`
	Brand      string         `json:"brand,omitempty"`
	Source     string         `json:"source"`
	Grams      float64        `json:"grams"`       // количество в граммах
	Nutrition  NutritionInfo  `json:"nutrition"`   // на 100г (для пересчёта)
	ImageURL   string         `json:"image_url,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	// Пересчитанные значения на указанное кол-во грамм (для ответа API)
	TotalCalories      float64 `json:"total_calories"`
	TotalProtein       float64 `json:"total_protein"`
	TotalFat           float64 `json:"total_fat"`
	TotalCarbohydrates float64 `json:"total_carbohydrates"`
}

// AddCartRequest — добавить продукт в корзину (тело запроса)
type AddCartRequest struct {
	Product FoodItem `json:"product"` // продукт из поиска/штрихкода
	Grams   float64  `json:"grams"`   // количество в граммах
}

// CartListResponse — список корзины с итогами
type CartListResponse struct {
	Items   []CartItem `json:"items"`
	TotalCalories      float64 `json:"total_calories"`
	TotalProtein       float64 `json:"total_protein"`
	TotalFat           float64 `json:"total_fat"`
	TotalCarbohydrates float64 `json:"total_carbohydrates"`
}

// UpdateCartRequest — изменить граммовку
type UpdateCartRequest struct {
	Grams float64 `json:"grams"`
}
