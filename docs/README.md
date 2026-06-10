# Documentación — Tenant Service

## Architecture Decision Records

| ADR | Título | Estado | Fecha |
|-----|--------|--------|-------|
| [ADR-001](adr/ADR-001-arquitectura-hexagonal-ddd.md) | Arquitectura Hexagonal y DDD | Aceptado | 2026-02-03 |
| [ADR-002](adr/ADR-002-separacion-tenant-db-eventbus.md) | Separación tenant_db y eventbus en PostgreSQL | Aceptado | 2026-02-03 |
| [ADR-003](adr/ADR-003-domain-events-via-eventbus.md) | Domain Events vía EventBus | Aceptado | 2026-02-19 |

## Arquitectura

- [Visión general](architecture/overview.md) — Capas, componentes, regla de dependencia, modelo de datos
- [Write Path](architecture/write-path.md) — Endpoints de escritura, optimistic locking, eventos de dominio

## Setup

- [Getting Started](setup/getting-started.md) — Variables de entorno, cómo levantar, correr tests

## Runbooks

- [Deployment](runbooks/deployment.md) — Cómo hacer deploy, migraciones, rollback

## Guías

- [Kong Integration](guides/kong-integration.md) — Rutas registradas, plugins, consumo desde otros servicios
- [Bootstrap Endpoint](guides/bootstrap-endpoint.md) — Inicialización de tenant, integración con onboarding
