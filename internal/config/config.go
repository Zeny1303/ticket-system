package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
// Every other package reads config from here — never from os.Getenv directly.
// This is the Single Responsibility Principle applied to configuration.
type Config struct {
	// Server
	Port string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret      string
	JWTExpiryHours int
}

// Load reads the .env file and populates the Config struct.
// It is called once at application startup in main.go.
// If a required variable is missing, the app logs a fatal error and exits.
// This is intentional — a misconfigured app should fail fast, not silently.
func Load() *Config {
	// godotenv.Load() reads the .env file from the current directory.
	// If the file doesn't exist (e.g., in production where env vars are
	// injected directly), we log a warning but continue — the app might
	// still work with environment variables set by the deployment platform.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found. Reading from environment variables.")
	}

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "ticketdb"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
	}

	// JWT secret is non-negotiable — the app cannot work without it.
	// An empty JWT secret would mean all tokens are signed with an empty
	// key, which is a critical security vulnerability.
	if cfg.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is required and cannot be empty.")
	}

	// Parse JWT expiry hours with a safe default of 24 hours.
	expiryStr := getEnv("JWT_EXPIRY_HOURS", "24")
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil || expiry <= 0 {
		log.Printf("Warning: Invalid JWT_EXPIRY_HOURS '%s'. Defaulting to 24 hours.", expiryStr)
		expiry = 24
	}
	cfg.JWTExpiryHours = expiry

	return cfg
}

// DSN (Data Source Name) builds the PostgreSQL connection string from config fields.
// GORM's postgres driver requires this exact format.
// Example output: "host=localhost user=postgres password=secret dbname=ticketdb port=5432 sslmode=disable TimeZone=UTC"
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort,
	)
}

// getEnv is a helper that reads an environment variable.
// If the variable is not set, it returns the provided fallback value.
// This pattern is cleaner than repeating os.Getenv + nil checks everywhere.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}