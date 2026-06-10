# Getting Started — Tenant Service

## Requisitos previos

- Docker + docker compose
- Infra del lab corriendo (`make -C ~/Projects infra`)
- Go 1.24+ (para desarrollo sin Docker)

## Variables de entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `DB_HOST` | Host de PostgreSQL | `lab-postgres` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_USER` | Usuario de PostgreSQL | `postgres` |
| `DB_PASSWORD` | Contraseña de PostgreSQL | `postgres` |
| `DB_NAME` | Nombre de la base de datos | `tenant_db` |
| `EVENTBUS_DB_HOST` | Host del EventBus DB | `lab-postgres` |
| `EVENTBUS_DB_NAME` | Nombre de la base del EventBus | `eventbus` |
| `PORT` | Puerto del servidor HTTP | `8120` |
| `PROMETHEUS_ENABLED` | Exponer `/metrics` | `true` |
| `GIN_MODE` | Modo de Gin (`debug`/`release`) | `debug` |
| `JWT_SECRET` | Secreto para validar JWT | *(requerido en prod)* |

## Levantar en local con Docker

```bash
# 1. Levantar infra compartida (si no está corriendo)
make -C ~/Projects infra

# 2. Levantar el tenant-service
docker compose up -d

# 3. Verificar que está sano
curl http://localhost:8125/health
```

El servicio queda disponible en `localhost:8125`.

## Levantar sin Docker (desarrollo Go directo)

```bash
# Variables de entorno
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=tenant_db
export EVENTBUS_DB_HOST=localhost EVENTBUS_DB_NAME=eventbus
export PORT=8120 PROMETHEUS_ENABLED=true JWT_SECRET=dev-secret

# Correr
go run ./cmd/api
```

## Correr tests

```bash
go test ./...
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
```

## Verificar que funciona

```bash
# Health check
curl http://localhost:8125/health

# Métricas Prometheus
curl http://localhost:8125/metrics | grep "go_info"

# Via Kong Gateway
curl http://localhost:8000/tenant/health
```
