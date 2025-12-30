package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Server
	HTTPPort int
	GRPCPort int

	// Database
	PostgresURL string
	MongoURL    string
	MongoDB     string

	// Security
	JWTSecret string

	// Environment
	Environment string
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		HTTPPort:    getEnvInt("HTTP_PORT", 8080),
		GRPCPort:    getEnvInt("GRPC_PORT", 8081),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://devjournal:devjournal_secret@localhost:5432/devjournal?sslmode=disable"),
		MongoURL:    getEnv("MONGO_URL", "mongodb://devjournal:devjournal_secret@localhost:27017"),
		MongoDB:     getEnv("MONGO_DB", "devjournal"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
