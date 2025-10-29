# Makefile for Microservice Demo

.PHONY: help build up down logs test setup clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all Docker images
	@echo "Building Docker images..."
	docker-compose build

up: ## Start all services
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 30
	@echo "Services are ready!"

down: ## Stop all services
	@echo "Stopping all services..."
	docker-compose down

logs: ## Show logs for all services
	docker-compose logs -f

logs-service1: ## Show logs for service1 only
	docker-compose logs -f service1

logs-service2: ## Show logs for service2 only
	docker-compose logs -f service2

setup: ## Setup RabbitMQ after starting services
	@echo "Setting up RabbitMQ..."
	./setup-rabbitmq.sh

test: ## Run tests against the services
	@echo "Running tests..."
	./test-services.sh

status: ## Check service status
	@echo "Service Status:"
	docker-compose ps

health: ## Check health of all services
	@echo "Checking service health..."
	@echo "Order Service (should return 200):"
	@curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/health || echo "Service not available"
	@echo "Report Service (should return 200):"
	@curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8081/health || echo "Service not available"

rabbitmq-ui: ## Open RabbitMQ management UI
	@echo "Opening RabbitMQ Management UI..."
	@echo "URL: http://localhost:15672"
	@echo "Username: guest"
	@echo "Password: guest"

jaeger-ui: ## Open Jaeger UI
	@echo "Opening Jaeger UI..."
	@echo "URL: http://localhost:16686"

clean: ## Clean up containers, volumes and images
	@echo "Cleaning up..."
	docker-compose down -v --rmi all
	docker system prune -f

dev-service1: ## Run service1 in development mode
	@echo "Starting service1 in development mode..."
	cd service1 && go run main.go

dev-service2: ## Run service2 in development mode
	@echo "Starting service2 in development mode..."
	cd service2 && go run main.go

install-deps-service1: ## Install dependencies for service1
	cd service1 && go mod tidy && go mod download

install-deps-service2: ## Install dependencies for service2
	cd service2 && go mod tidy && go mod download

install-deps: install-deps-service1 install-deps-service2 ## Install dependencies for both services

full-setup: build up setup ## Build, start services and setup RabbitMQ
	@echo "Full setup completed!"
	@echo "Services available at:"
	@echo "  - Order Service: http://localhost:8080"
	@echo "  - Report Service: http://localhost:8081"
	@echo "  - RabbitMQ UI: http://localhost:15672"
	@echo "  - Jaeger UI: http://localhost:16686"