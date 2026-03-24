package store

import (
	"context"
	"database/sql"

	"calorie-proxy/internal/models"
)

// AddCartItem добавляет продукт в корзину пользователя (в граммах)
func (s *Store) AddCartItem(ctx context.Context, userID int64, product models.FoodItem, grams float64) (*models.CartItem, error) {
	n := product.Nutrition
	var id int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO cart_items (user_id, food_id, name, brand, source, grams,
			calories_per_100g, protein_per_100g, fat_per_100g, carbs_per_100g, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		userID, product.ID, product.Name, product.Brand, product.Source, grams,
		n.Calories, n.Protein, n.Fat, n.Carbohydrates, product.ImageURL,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.getCartItemByID(ctx, id, userID)
}

// GetCartByUserID возвращает все позиции корзины с пересчитанными калориями/БЖУ на граммовку
func (s *Store) GetCartByUserID(ctx context.Context, userID int64) ([]models.CartItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, food_id, name, brand, source, grams,
			calories_per_100g, protein_per_100g, fat_per_100g, carbs_per_100g, image_url, created_at
		 FROM cart_items WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.CartItem
	for rows.Next() {
		var c models.CartItem
		var cal, pr, fat, carb float64
		err := rows.Scan(&c.ID, &c.UserID, &c.FoodID, &c.Name, &c.Brand, &c.Source, &c.Grams,
			&cal, &pr, &fat, &carb, &c.ImageURL, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		c.Nutrition = models.NutritionInfo{Calories: cal, Protein: pr, Fat: fat, Carbohydrates: carb}
		coef := c.Grams / 100
		c.TotalCalories = cal * coef
		c.TotalProtein = pr * coef
		c.TotalFat = fat * coef
		c.TotalCarbohydrates = carb * coef
		items = append(items, c)
	}
	return items, rows.Err()
}

// UpdateCartItemGrams обновляет граммовку позиции
func (s *Store) UpdateCartItemGrams(ctx context.Context, itemID, userID int64, grams float64) (*models.CartItem, error) {
	_, err := s.db.ExecContext(ctx,
		`UPDATE cart_items SET grams = $1 WHERE id = $2 AND user_id = $3`,
		grams, itemID, userID,
	)
	if err != nil {
		return nil, err
	}
	return s.getCartItemByID(ctx, itemID, userID)
}

// DeleteCartItem удаляет позицию из корзины
func (s *Store) DeleteCartItem(ctx context.Context, itemID, userID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM cart_items WHERE id = $1 AND user_id = $2`, itemID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) getCartItemByID(ctx context.Context, id, userID int64) (*models.CartItem, error) {
	var c models.CartItem
	var cal, pr, fat, carb float64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, food_id, name, brand, source, grams,
			calories_per_100g, protein_per_100g, fat_per_100g, carbs_per_100g, image_url, created_at
		 FROM cart_items WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&c.ID, &c.UserID, &c.FoodID, &c.Name, &c.Brand, &c.Source, &c.Grams,
		&cal, &pr, &fat, &carb, &c.ImageURL, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Nutrition = models.NutritionInfo{Calories: cal, Protein: pr, Fat: fat, Carbohydrates: carb}
	coef := c.Grams / 100
	c.TotalCalories = cal * coef
	c.TotalProtein = pr * coef
	c.TotalFat = fat * coef
	c.TotalCarbohydrates = carb * coef
	return &c, nil
}
