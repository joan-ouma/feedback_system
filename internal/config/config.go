package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database      DatabaseConfig
	Server        ServerConfig
	LLM           LLMConfig
	OpenTelemetry OpenTelemetryConfig
}

type DatabaseConfig struct {
	URI string // MongoDB connection URI
}

type ServerConfig struct {
	Port         string
	SessionSecret string
}

type LLMConfig struct {
	APIURL string
	APIKey string
	Model  string
}

type OpenTelemetryConfig struct {
	JaegerEndpoint string
	Enabled        bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		Database: DatabaseConfig{
			URI: getEnv("MONGODB_URI", getEnv("DATABASE_URL", "mongodb://localhost:27017/feedback_sys")), // Support both for migration
		},
		Server: ServerConfig{
			Port:          getEnv("PORT", "8080"),
			SessionSecret: getEnv("SESSION_SECRET", "change-me-in-production"),
		},
		LLM: LLMConfig{
			APIURL: getEnv("LLM_API_URL", "https://api.openai.com/v1"),
			APIKey: getEnv("LLM_API_KEY", ""),
			Model:  getEnv("LLM_MODEL", "gpt-4"),
		},
		OpenTelemetry: OpenTelemetryConfig{
			JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			Enabled:        getEnvAsBool("ENABLE_TRACING", true),
		},
	}

	// LLM_API_KEY is optional - consultation feature will show a message if not configured
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

