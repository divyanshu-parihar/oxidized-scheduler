package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	AppEnv      string
}

func LoadConfig() *Config {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// Try loading .env based on APP_ENV
	// e.g., .env.production, .env.development
	envFile := ".env." + appEnv
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			slog.Warn("Error loading env file", "file", envFile, "error", err)
		} else {
			slog.Info("Loaded configuration", "file", envFile)
		}
	} else {
		// Fallback to default .env
		if err := godotenv.Load(); err != nil {
			slog.Warn("No .env file found, using system environment variables")
		} else {
			slog.Info("Loaded configuration from default .env")
		}
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/scheduler?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		AppEnv:      appEnv,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
