package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
	DB              DB
}

type DB struct {
	URL      string
	MaxConns int32
	MinConns int32
}

func Load() (Config, error) {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("config: load .env: %w", err)
	}

	cfg := Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		ShutdownTimeout: 10 * time.Second,
		DB: DB{
			URL:      os.Getenv("DATABASE_URL"),
			MaxConns: 10,
			MinConns: 2,
		},
	}

	if cfg.DB.URL == "" {
		return Config{}, errors.New("config: DATABASE_URL is required")
	}
	return cfg, nil
}

// getEnv returns the value of the environment variable named by the key.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
