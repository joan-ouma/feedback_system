package config

import (
	"log"
	"os"
	"strconv"
	"strings"

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

	// Debug: Log environment variables (masked)
	mongoURI := getEnv("MONGODB_URI", "")
	if mongoURI == "" {
		mongoURI = getEnv("DATABASE_URL", "")
	}
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/feedback_sys" // Default fallback
		log.Println("⚠️  WARNING: MONGODB_URI not set, using default localhost")
	} else {
		// Mask password in log
		maskedURI := maskURI(mongoURI)
		log.Printf("✅ Using MongoDB URI: %s", maskedURI)
	}

	cfg := &Config{
		Database: DatabaseConfig{
			URI: mongoURI,
		},
		Server: ServerConfig{
			Port:          getEnv("PORT", "8080"),
			SessionSecret: getEnv("SESSION_SECRET", "change-me-in-production"),
		},
		LLM: LLMConfig{
			APIURL: getEnv("LLM_API_URL", getEnv("GEMINI_API_URL", "https://generativelanguage.googleapis.com/v1beta")), // Default to Gemini API
			APIKey: getEnv("LLM_API_KEY", getEnv("GEMINI_API_KEY", "")), // Only API key needed
			Model:  "", // Model not used - hardcoded in client
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

// maskURI masks password in MongoDB connection string for logging
func maskURI(uri string) string {
	// Format: mongodb+srv://user:password@host
	if idx := strings.Index(uri, "@"); idx > 0 {
		if userPassIdx := strings.Index(uri, "://"); userPassIdx > 0 {
			prefix := uri[:userPassIdx+3]
			userPass := uri[userPassIdx+3:idx]
			if colonIdx := strings.Index(userPass, ":"); colonIdx > 0 {
				user := userPass[:colonIdx]
				return prefix + user + ":***@" + uri[idx+1:]
			}
		}
	}
	return uri
}

