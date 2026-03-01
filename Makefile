.PHONY: help build run test clean docker-up docker-down docker-logs migrate

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compila el servicio
	@echo "🔨 Compilando tenant-service..."
	go build -o tenant-service ./cmd/api

run: ## Ejecuta el servicio localmente
	@echo "🚀 Iniciando tenant-service..."
	go run ./cmd/api/main.go

test: ## Ejecuta los tests
	@echo "🧪 Ejecutando tests..."
	go test -v ./...

test-coverage: ## Ejecuta tests con coverage
	@echo "📊 Ejecutando tests con coverage..."
	go test -cover -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Coverage report generado en coverage.html"

clean: ## Limpia archivos generados
	@echo "🧹 Limpiando..."
	rm -f tenant-service
	rm -f coverage.txt coverage.html
	rm -rf tmp/

docker-up: ## Inicia servicios con Docker Compose
	@echo "🐳 Iniciando servicios..."
	docker-compose up -d

docker-down: ## Para servicios Docker
	@echo "🛑 Parando servicios..."
	docker-compose down

docker-logs: ## Muestra logs de Docker
	docker-compose logs -f tenant-service

docker-rebuild: ## Reconstruye imagen Docker
	@echo "🔄 Reconstruyendo imagen..."
	docker-compose build --no-cache

migrate: ## Ejecuta migraciones manualmente
	@echo "📦 Ejecutando migraciones..."
	psql -h localhost -p 5435 -U postgres -d tenant_db -f migrations/001_create_tenant_config_table.sql
	psql -h localhost -p 5435 -U postgres -d tenant_db -f migrations/002_seed_initial_data.sql

deps: ## Descarga dependencias
	@echo "📥 Descargando dependencias..."
	go mod download
	go mod verify

lint: ## Ejecuta linter
	@echo "🔍 Ejecutando linter..."
	golangci-lint run

fmt: ## Formatea código
	@echo "✨ Formateando código..."
	go fmt ./...

tidy: ## Limpia dependencias no usadas
	@echo "🧹 Limpiando dependencias..."
	go mod tidy

.DEFAULT_GOAL := help
