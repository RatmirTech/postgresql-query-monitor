package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	// Vault configuration
	VaultPath  string // Path in Vault for our database credentials (where we store review results)
	VaultToken string
	VaultAddr  string

	// Review API configuration
	ReviewAPI struct {
		URL string
	}

	// Logging configuration
	LogPath string

	// Environment
	Environment string
}

// Load loads configuration from environment variables
func (c *Config) Load() error {
	// Vault configuration
	c.VaultToken = getEnv("VAULT_TOKEN", "root")
	c.VaultAddr = getEnv("VAULT_ADDR", "http://localhost:8200")

	// Review API
	c.ReviewAPI.URL = getEnv("REVIEW_API_URL", "http://")

	// Logging
	if c.LogPath == "" {
		c.LogPath = getEnv("PG_LOG_PATH", "/var/log/postgresql")
	}

	// Environment
	if c.Environment == "" {
		c.Environment = getEnv("ENVIRONMENT", "production")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
