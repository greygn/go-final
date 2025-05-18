package config

import (
	"os"
	"time"
)

type Config struct {
	PostgresURL     string
	GRPCAddr        string
	JWTSecretKey    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() *Config {
	return &Config{
		PostgresURL:     getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"),
		GRPCAddr:        getEnv("GRPC_ADDR", ":50051"),
		JWTSecretKey:    getEnv("JWT_SECRET", "your-secret-key"),
		AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_EXPIRY", "15m")),
		RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "168h")), // 7 days
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return time.Minute * 15 // default to 15 minutes
	}
	return duration
}
