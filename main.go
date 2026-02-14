package main

import (
	"context"
	api "github.com/divyanshu-parihar/oxidized-scheduler/api"
	"github.com/divyanshu-parihar/oxidized-scheduler/internal/config"
	"github.com/divyanshu-parihar/oxidized-scheduler/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
)

func main() {
	cfg := config.LoadConfig()

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		slog.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Connected to PostgreSQL", "env", cfg.AppEnv)

	a := api.NewAPI(db)
	slog.Info("Starting the server", "port", cfg.Port)
	if err := a.CreateServer().Run(":" + cfg.Port); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
