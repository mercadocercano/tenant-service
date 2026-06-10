# AUDIT.md — Tenant Service
**Fecha**: 2026-06-10 | **Auditor**: go-hex-audit pipeline

---

## Resumen ejecutivo

| Fase | Estado |
|------|--------|
| 0 Discovery | PASS — arquitectura hexagonal confirmada |
| 1 Compile | PASS — `go build ./...` sin errores |
| 2 Architecture audit | PASS — sin violaciones CRITICAL/HIGH |
| 3 Coverage | PASS — domain/application ≥90%, persistence 85.6% |
| 4 OpenAPI | PASS — spec completo, 7 endpoints documentados |
| 5 Postman collection | DONE — `postman/collection.json` creado |
| 6 E2E script | DONE — `scripts/e2e.sh` (requiere JWT_TOKEN + lab-postgres) |

---

## Fase 1 — Compile

```
GOWORK=off go build ./... → OK
GOWORK=off go vet ./...   → OK
```

**Nota preexistente**: El Docker development stage falla al resolver módulos privados
(`github.com/mercadocercano/eventbus`, `github.com/mercadocercano/middleware`) sin
`GITHUB_TOKEN`. Los módulos están en el cache local del host y los tests locales pasan.
Solución: agregar `GITHUB_TOKEN` como `build_arg` en docker-compose.yml o usar un
registry privado. No introducido por este audit.

---

## Fase 2 — Hexagonal/DDD audit

Layering verificado con `go list -f '{{.ImportPath}}...' ./src/...`

| # | Severidad | Archivo | Hallazgo | Impacto |
|---|-----------|---------|----------|---------|
| 1 | LOW | `application/command/*.go` | Import de `log` en capa application | La capa application no debería emitir logs directos; debería retornar errores que el adapter loguea |
| 2 | LOW | `application/response/*.go` | Import de `domain/entity` para conversión DTO | Patrón aceptable (assembler pattern); los DTOs de respuesta necesitan el tipo de dominio |

**Sin violaciones CRITICAL o HIGH.** La dependencia `domain ← application ← infrastructure` se respeta en todos los paquetes. Los repositorios se definen como interfaces en `domain/repository/` y sus implementaciones en `infrastructure/persistence/` importan correctamente la interfaz, no al revés.

---

## Fase 3 — Coverage

| Package | Coverage | Umbral | Estado |
|---------|----------|--------|--------|
| `domain/entity` | 100% | 90% | PASS |
| `domain/exception` | 100% | 90% | PASS |
| `domain/valueobject` | 91.7% | 90% | PASS |
| `application/command` | 94.9% | 90% | PASS |
| `application/query` | 100% | 90% | PASS |
| `infrastructure/persistence` | 85.6% | 80% | PASS |

Tests escritos en este audit:
- `postgres_tenant_settings_repository_test.go` — 7 tests (GetByTenantID, Save insert/update/version_conflict, Exists, DB error)
- `postgres_point_of_sale_repository_test.go` — 8 tests (Create, GetByID found/not_found, ListByTenant empty, ListActiveByTenant, Update found/not_found)
- Completado `postgres_tenant_config_repository_test.go` — agregados Delete, GetAllByTenant, GetAllByTenant_Empty

Deuda de coverage (no en threshold, controllers/event/shared = 0%):
- `infrastructure/controller/` — requiere mocks HTTP; out of scope para este audit unitario
- `infrastructure/event/` — requiere mock del eventbus
- `shared/infrastructure/config/` — trivial, no hay lógica de negocio

---

## Fase 4 — OpenAPI

Spec `api-docs/openapi.yaml` completo. Endpoints verificados:

| Método | Path | Documentado |
|--------|------|-------------|
| GET | `/tenant/config/{key}` | ✓ |
| POST | `/tenant/config` | ✓ |
| POST | `/tenant/bootstrap` | ✓ |
| GET | `/tenant/settings` | ✓ |
| PUT | `/tenant/settings` | ✓ |
| GET | `/tenant/points-of-sale` | ✓ |
| POST | `/tenant/points-of-sale` | ✓ |

Sin cambios requeridos al spec.

---

## Fase 5 — Postman collection

`postman/collection.json` creado (v2.1). Incluye:
- Variables: `{{baseUrl}}`, `{{tenantId}}`, `{{jwtToken}}`
- 9 requests organizados en carpetas: Health, Bootstrap, Tenant Config, Tenant Settings, Points of Sale, Casos negativos
- Tests de assertions en cada request
- `postman/environment.local.json` para `localhost:8125`

---

## Fase 6 — E2E script

`scripts/e2e.sh` creado. Requiere:
- `mc-tenant-service` corriendo contra `lab-postgres` (`docker compose up -d`)
- `JWT_TOKEN=<token>` para tests autenticados

Estrategia: PostgreSQL (no SQLite) porque el servicio usa JSONB y ON CONFLICT específicos de PostgreSQL.

---

## Deuda pendiente

| Item | Razón |
|------|-------|
| Docker dev build | Requiere GITHUB_TOKEN para módulos privados — configurar como `build_arg` o usar un registry privado |
| Coverage controllers | Requiere mocks HTTP — aumentaría el suite pero no está en el critical path |
| `go list` LOW findings | No bloqueantes; resolver en una tarea de cleanup separada |
| newman e2e automatizado en CI | Requiere JWT_TOKEN disponible en el entorno CI — pendiente de configuración |
