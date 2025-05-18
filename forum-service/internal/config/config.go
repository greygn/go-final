package config

import (
	"os"
	"time"
)

type Config struct {
	PostgresURL      string
	GRPCAddr         string
	HTTPAddr         string
	AuthServiceAddr  string
	MessageTTL       time.Duration
	WebSocketTimeout time.Duration
}

func Load() *Config {
	return &Config{
		PostgresURL:      getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/forum?sslmode=disable"),
		GRPCAddr:         getEnv("GRPC_ADDR", ":50052"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		AuthServiceAddr:  getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		MessageTTL:       time.Second * 20,
		WebSocketTimeout: time.Second * 60,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
