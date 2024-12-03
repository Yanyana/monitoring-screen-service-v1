package database

import (
	"context"
	"go-service/config"
	"log"

	"github.com/go-redis/redis/v8"
)

func InitRedis(cfg *config.ConfigStruc) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}
