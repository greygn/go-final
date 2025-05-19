package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL     string
	JWTSecretKey    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	GRPCAddr        string
	HTTPAddr        string
}

func Load() (*Config, error) {
	config := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"),
		JWTSecretKey:    getEnv("JWT_SECRET_KEY", "your-secret-key"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		GRPCAddr:        getEnv("GRPC_ADDR", ":50051"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8082"),
	}
	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
