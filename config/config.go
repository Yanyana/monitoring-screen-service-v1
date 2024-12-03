package config

import (
	"log"

	"github.com/spf13/viper"
)

type ConfigStruc struct {
	ServerPort  string
	PostgresDSN string
	RedisAddr   string
	RedisPass   string
}

func LoadConfig() *ConfigStruc {
	// Set up Viper to read from .env file
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("No .env file found, relying on environment variables: %v", err)
	}

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASS", "")

	// Read environment variables
	viper.AutomaticEnv()

	cfg := &ConfigStruc{
		ServerPort:  viper.GetString("SERVER_PORT"),
		PostgresDSN: viper.GetString("POSTGRES_DSN"),
		RedisAddr:   viper.GetString("REDIS_ADDR"),
		RedisPass:   viper.GetString("REDIS_PASS"),
	}

	if cfg.PostgresDSN == "" {
		log.Fatalf("Postgres DSN is required")
	}

	return cfg
}
