package database

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string, migrationsPath string) error {
	// Transform postgres:// to pgx5:// for golang-migrate
	migrationURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		migrationURL,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}
