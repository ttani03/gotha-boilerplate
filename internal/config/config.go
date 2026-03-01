// Package config provides application configuration management.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	Port                    string
	Env                     string
	DatabaseURL             string
	JWTSecret               string `json:"-"`
	JWTAccessTokenDuration  time.Duration
	JWTRefreshTokenDuration time.Duration
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	accessDuration, err := time.ParseDuration(getEnv("JWT_ACCESS_TOKEN_DURATION", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_DURATION: %w", err)
	}

	refreshDuration, err := time.ParseDuration(getEnv("JWT_REFRESH_TOKEN_DURATION", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TOKEN_DURATION: %w", err)
	}

	return &Config{
		Port:                    port,
		Env:                     env,
		DatabaseURL:             databaseURL,
		JWTSecret:               jwtSecret,
		JWTAccessTokenDuration:  accessDuration,
		JWTRefreshTokenDuration: refreshDuration,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
