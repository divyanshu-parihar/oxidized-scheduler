include .env.development
export $(shell sed 's/=.*//' .env.development)

CONTAINER_TOOL ?= podman
COMPOSE_TOOL ?= podman compose

.PHONY: db-up db-down migrate-up migrate-down seed db-clean dashboard test help

help:
	@echo "Available commands:"
	@echo "  db-up        - Start containers using $(COMPOSE_TOOL)"
	@echo "  db-down      - Stop and remove containers"
	@echo "  db-clean     - Truncate all tasks in the DB"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  seed         - Seed the database with randomized events"
	@echo "  dashboard    - Start the admin dashboard"
	@echo "  test         - Run all tests"

db-up:
	$(COMPOSE_TOOL) up -d

db-down:
	$(COMPOSE_TOOL) down

migrate-up:
	go run cmd/migrate/main.go -cmd up

migrate-down:
	go run cmd/migrate/main.go -cmd down

db-clean:
	$(COMPOSE_TOOL) exec -T db psql -U postgres -d scheduler -c "TRUNCATE tasks;"

seed:
	go run cmd/seed/main.go

dashboard:
	go run cmd/dashboard/main.go

test:
	go test -v ./...
