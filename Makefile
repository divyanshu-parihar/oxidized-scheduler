include .env.development
export $(shell sed 's/=.*//' .env.development)

.PHONY: db-up db-down migrate-up migrate-down seed help

help:
	@echo "Available commands:"
	@echo "  db-up        - Start PostgreSQL container"
	@echo "  db-down      - Stop and remove PostgreSQL container"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  seed         - Seed the database with 50k events"

db-up:
	docker-compose up -d

db-down:
	docker-compose down

migrate-up:
	go run cmd/migrate/main.go -cmd up

migrate-down:
	go run cmd/migrate/main.go -cmd down

seed:
	go run cmd/seed/main.go
