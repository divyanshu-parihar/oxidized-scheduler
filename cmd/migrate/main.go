package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/divyanshu-parihar/oxidized-scheduler/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.LoadConfig()

	var command string
	flag.StringVar(&command, "cmd", "up", "Migration command (up, down, version)")
	flag.Parse()

	migrationURL := strings.Replace(cfg.DatabaseURL, "postgres://", "pgx5://", 1)

	m, err := migrate.New(
		"file://migrations",
		migrationURL,
	)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("could not run up migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("could not run down migrations: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("could not get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
