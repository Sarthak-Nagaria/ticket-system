package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port           string
	DatabasePath   string
	JWTSecret      string
	JWTExpiryHours int
	GinMode        string
}

// Load reads configuration from a .env file (if present) and environment
// variables, falling back to sane defaults. Environment variables always
// take precedence over values in .env.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables/defaults")
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabasePath:   getEnv("DATABASE_PATH", "ticket_system.db"),
		JWTSecret:      getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		GinMode:        getEnv("GIN_MODE", "release"),
	}

	if cfg.JWTSecret == "change-this-secret-in-production" {
		log.Println("WARNING: using default JWT secret. Set JWT_SECRET in production.")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	var result int
	for _, c := range v {
		if c < '0' || c > '9' {
			return fallback
		}
	}
	result = 0
	for _, c := range v {
		result = result*10 + int(c-'0')
	}
	return result
}
