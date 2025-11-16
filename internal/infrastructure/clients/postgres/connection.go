package postgres

import (
	"fmt"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func CreateConnection(cfg *config.PostgresConfig) (*sqlx.DB, error) {
	ds := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.Host,
		cfg.Port,
		"disable",
	)

	db, err := sqlx.Connect("postgres", ds)

	if err != nil {
		return nil, fmt.Errorf("failed connection to db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}
