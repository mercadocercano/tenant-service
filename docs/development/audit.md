# Auditoría Hexagonal/DDD — Tenant Service

Fecha: 2026-06-10 | Auditor: go-hex-audit skill

---

## Resumen ejecutivo

| Fase | Estado | Detalle |
|------|--------|---------|
| 0 Discovery | ✅ PASS | Arquitectura hexagonal confirmada |
| 1 Compile | ✅ PASS | 11 archivos formateados con gofmt |
| 2 Architecture audit | ✅ PASS | Sin violations CRITICAL/HIGH; 2 findings LOW |
| 3 Coverage | ✅ PASS | 47.4% → **73.1%** (+25.7 pp); domain/app >90% |
| 4 OpenAPI | ✅ PASS | Spec completa; endpoint `/health` agregado; 0 errores lint |
| 5 Postman collection | ✅ PASS | Colección preexistente cubre todos los endpoints |
| 6 E2E | ⚠️ SKIP | Script existe; requiere lab-postgres levantado + JWT_TOKEN |

---

## Fase 1 — Correcciones de compilación

**gofmt aplicado** a 11 archivos:
- `src/tenant/application/command/bootstrap_tenant_config_command_test.go`
- `src/tenant/application/command/set_tenant_config_command_test.go`
- `src/tenant/application/command/update_tenant_settings_command.go`
- `src/tenant/application/command/update_tenant_settings_command_test.go`
- `src/tenant/application/query/query_test.go`
- `src/tenant/application/request/create_point_of_sale_request.go`
- `src/tenant/domain/entity/point_of_sale.go`
- `src/tenant/domain/entity/tenant_settings.go`
- `src/tenant/infrastructure/config/tenant_config.go`
- `src/tenant/infrastructure/controller/tenant_config_controller.go`
- `src/tenant/infrastructure/persistence/postgres_tenant_config_repository_test.go`

---

## Fase 2 — Auditoría de acoplamiento hexagonal

**Grafo de dependencias** (sin violations):
```
domain/entity, domain/repository, domain/valueobject, domain/exception
  <- application/command, application/query, application/request, application/response
  <- infrastructure/controller, infrastructure/persistence, infrastructure/event, infrastructure/config
  <- cmd/api
```

**Hallazgos:**

| # | Severity | Archivo | Violación | Sugerencia |
|---|----------|---------|-----------|------------|
| 1 | LOW | `domain/entity/tenant_settings.go` | `allowed_currencies` se serializa via `json.Marshal` en el repositorio — responsabilidad de serialización en infrastructure ✅ | No action needed |
| 2 | LOW | `application/command/update_tenant_settings_command.go` | `encoding/json` para construir el payload del evento directamente sin DTO explícito | Considerar tipo `TenantSettingsUpdatedPayload` para tipar el payload de evento |

---

## Fase 3 — Coverage

### Antes / Después

| Package | Antes | Después |
|---------|-------|---------|
| domain/entity | 100% | 100% |
| domain/exception | 100% | 100% |
| domain/valueobject | 91.7% | 91.7% |
| application/command | 94.9% | 94.9% |
| application/query | 100% | 100% |
| application/request | **0%** | **100%** |
| application/response | **0%** | **100%** |
| infrastructure/controller | **0%** | **79.4%** |
| infrastructure/persistence | 85.6% | 85.6% |
| **Total** | **47.4%** | **73.1%** |

### Tests nuevos creados

- `src/tenant/application/request/request_test.go` — 14 tests; Validate() en 3 tipos de request
- `src/tenant/application/response/response_test.go` — 7 tests; mapeo de DTOs
- `src/tenant/infrastructure/controller/tenant_config_controller_test.go` — 12 tests
- `src/tenant/infrastructure/controller/tenant_settings_controller_test.go` — 8 tests
- `src/tenant/infrastructure/controller/point_of_sale_controller_test.go` — 9 tests

### Packages en 0% (justificados)

| Package | Razón |
|---------|-------|
| `cmd/api` | Wiring del servidor; no tiene lógica testeable unitariamente |
| `infrastructure/config` | DI wiring; instancia use cases |
| `infrastructure/event` | Thin adapter sobre eventbus; lógica en la lib |
| `shared/infrastructure/config` | Config CORS de 2 funciones; comportamiento en gin |

---

## Fase 4 — OpenAPI

**Endpoints agregados/corregidos:**
- Agregado `GET /health` (endpoint público, sin auth)
- Corregido `SimpleConfigResponse.value`: `nullable: true` → `type: ["string", "null"]` (sintaxis OpenAPI 3.1)
- Agregado `license.identifier: MIT`

**Resultado lint:** ✅ 0 errores, 2 warnings (localhost servers — esperados en spec de dev)

---

## Fase 5 — Postman collection

Colección preexistente en `postman/collection.json` cubre:
- Health check (sin auth)
- Bootstrap (idempotente)
- Tenant config key-value (set + get)
- Tenant settings (get + update)
- Points of sale (create + list)
- Casos negativos (sin token, sin X-Tenant-ID)

Variables: `{{baseUrl}}`, `{{tenantId}}`, `{{jwtToken}}`

---

## Fase 6 — E2E

**Por qué no SQLite:** el servicio usa `ON CONFLICT DO UPDATE` y serialización JSONB — no portable a SQLite.

**Script existente:** `scripts/e2e.sh` — apunta a `http://localhost:8125` (contenedor con lab-postgres).

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@test.com","password":"password"}' | jq -r .token)
JWT_TOKEN=$TOKEN bash scripts/e2e.sh
```

---

## Fix aplicado — go.mod / GOWORK=off (HIGH)

**Problema:** `main.go` importaba `github.com/hornosg/go-shared` (nuevo nombre de `libs/go-shared`) pero go.mod referenciaba `github.com/mercadocercano/go-shared v0.2.0` y `github.com/mercadocercano/middleware v0.1.0`. El build Docker (`GOWORK=off`) fallaba.

**Fix:** `GOWORK=off go mod tidy` actualizó go.mod:
- ✅ Agregó `github.com/hornosg/go-shared v0.3.0`
- 🗑 Eliminó `github.com/mercadocercano/go-shared v0.2.0`
- 🗑 Eliminó `github.com/mercadocercano/middleware v0.1.0`

**Resultado:** `GOWORK=off go build ./...` pasa ✅

---

## Deuda pendiente

| Item | Prioridad | Motivo |
|------|-----------|--------|
| Tests de SetConfig → error de repo (path no cubierto) | LOW | Coverage 66.7%; no es crítico |
| Tipo explícito para payload de evento en UpdateTenantSettingsCommand | LOW | Refactor, no bug |
| Coverage controller al 90% | MEDIUM | Los error paths restantes requieren mocks más complejos |
| E2E newman automático en CI | MEDIUM | Requiere lab-postgres en el pipeline |
