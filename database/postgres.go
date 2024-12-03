package database

import (
	"context"
	"go-service/config"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres(cfg *config.ConfigStruc) *pgxpool.Pool {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	return db
}
