package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Store обёртка над БД для пользователей и корзины
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// OpenDB открывает PostgreSQL и выполняет миграции
func OpenDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id),
			food_id TEXT NOT NULL,
			name TEXT NOT NULL,
			brand TEXT,
			source TEXT NOT NULL,
			grams DOUBLE PRECISION NOT NULL,
			calories_per_100g DOUBLE PRECISION NOT NULL,
			protein_per_100g DOUBLE PRECISION NOT NULL,
			fat_per_100g DOUBLE PRECISION NOT NULL,
			carbs_per_100g DOUBLE PRECISION NOT NULL,
			image_url TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_cart_user ON cart_items(user_id);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
