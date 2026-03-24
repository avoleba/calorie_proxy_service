package clients

import (
    "context"
    "calorie-proxy/internal/models"
)

type FoodDataProvider interface {
    SearchFoods(ctx context.Context, query string, page, pageSize int) ([]models.FoodItem, int, error)
    
    GetFoodByBarcode(ctx context.Context, barcode string) (*models.FoodItem, error)
    
    Name() string
}