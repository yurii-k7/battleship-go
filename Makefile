.PHONY: help build up down logs clean test backend-test frontend-test

# Default target
help:
	@echo "Available commands:"
	@echo "  build          - Build all Docker images"
	@echo "  up             - Start all services"
	@echo "  down           - Stop all services"
	@echo "  logs           - Show logs from all services"
	@echo "  clean          - Clean up Docker resources"
	@echo "  test           - Run all tests"
	@echo "  backend-test   - Run backend tests"
	@echo "  frontend-test  - Run frontend tests"
	@echo "  backend-deps   - Install backend dependencies"
	@echo "  frontend-deps  - Install frontend dependencies"

# Docker commands
build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Development commands
backend-deps:
	cd backend && go mod tidy

frontend-deps:
	cd frontend && npm install

# Test commands
test: backend-test frontend-test

backend-test:
	cd backend && go test ./...

frontend-test:
	cd frontend && npm test

# Database commands
db-reset:
	docker-compose down postgres
	docker volume rm battleship-go_postgres_data || true
	docker-compose up -d postgres

# Development setup
setup: backend-deps frontend-deps
	cp .env.example .env
	@echo "Setup complete! Edit .env file with your configuration."
	@echo "Run 'make up' to start the application."
