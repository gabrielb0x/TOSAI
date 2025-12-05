package config

import (
	"log"
	"os"
)

// Config holds configuration values loaded from environment variables.
type Config struct {
	Port           string
	DatabaseURL    string
	OpenAIAPIKey   string
	AdminPassword  string
	AllowedOrigins string
}

// Load reads environment variables into a Config struct.
func Load() Config {
	cfg := Config{
		Port:           getenvDefault("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
		AllowedOrigins: getenvDefault("ALLOWED_ORIGINS", "*"),
	}

	if cfg.DatabaseURL == "" {
		log.Println("warning: DATABASE_URL is not set; database connection will fail until configured")
	}

	return cfg
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
