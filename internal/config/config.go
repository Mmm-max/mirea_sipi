package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAppPort        = "8080"
	defaultAccessTTLMin   = 15
	defaultRefreshTTLHour = 720
)

type Config struct {
	App AppConfig
	DB  DBConfig
	JWT JWTConfig
}

type AppConfig struct {
	Port string
}

type DBConfig struct {
	DSN string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessTTLMin, err := envInt("JWT_ACCESS_TTL_MINUTES", defaultAccessTTLMin)
	if err != nil {
		return nil, fmt.Errorf("parse JWT_ACCESS_TTL_MINUTES: %w", err)
	}

	refreshTTLHour, err := envInt("JWT_REFRESH_TTL_HOURS", defaultRefreshTTLHour)
	if err != nil {
		return nil, fmt.Errorf("parse JWT_REFRESH_TTL_HOURS: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Port: env("APP_PORT", defaultAppPort),
		},
		DB: DBConfig{
			DSN: os.Getenv("POSTGRES_DSN"),
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			AccessTTL:  time.Duration(accessTTLMin) * time.Minute,
			RefreshTTL: time.Duration(refreshTTLHour) * time.Hour,
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	switch {
	case c.App.Port == "":
		return fmt.Errorf("APP_PORT is required")
	case c.DB.DSN == "":
		return fmt.Errorf("POSTGRES_DSN is required")
	case c.JWT.Secret == "":
		return fmt.Errorf("JWT_SECRET is required")
	case c.JWT.AccessTTL <= 0:
		return fmt.Errorf("JWT_ACCESS_TTL_MINUTES must be positive")
	case c.JWT.RefreshTTL <= 0:
		return fmt.Errorf("JWT_REFRESH_TTL_HOURS must be positive")
	default:
		return nil
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}
