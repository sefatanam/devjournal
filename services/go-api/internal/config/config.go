package config

import (
	"os"
	"strconv"
	"strings"
)

// @REVIEW: Simplified config with clear variable names
//
// Required Environment Variables:
//   PORT        - Server port (Railway sets automatically)
//   DB_URL      - PostgreSQL connection URL
//   MONGO_URL   - MongoDB connection URL
//   JWT_SECRET  - Secret for JWT tokens
//
// Optional:
//   GRPC_PORT   - gRPC server port (default: 8081)
//   MONGO_DB    - MongoDB database name (default: devjournal)

type Config struct {
	Port      int
	GRPCPort  int
	DbURL     string
	MongoURL  string
	MongoDB   string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnvInt("PORT", 8080),
		GRPCPort:  getEnvInt("GRPC_PORT", 8081),
		DbURL:     normalizeDbURL(getEnv("DB_URL", "postgres://devjournal:devjournal_secret@localhost:5432/devjournal?sslmode=disable")),
		MongoURL:  getEnv("MONGO_URL", "mongodb://devjournal:devjournal_secret@localhost:27017"),
		MongoDB:   getEnv("MONGO_DB", "devjournal"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
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

// normalizeDbURL handles both postgres:// and postgresql:// schemes
func normalizeDbURL(url string) string {
	return strings.Replace(url, "postgresql://", "postgres://", 1)
}
