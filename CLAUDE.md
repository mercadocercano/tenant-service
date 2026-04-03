# CLAUDE.md — Tenant Service

Centro de configuración operativa y fiscal del tenant: settings estructurados, puntos de venta y configuración clave-valor. Publica eventos para otros dominios.

**Puerto**: 8120 | **Stack**: Go + Gin + PostgreSQL + EventBus | **Arquitectura**: Hexagonal bajo `src/tenant/`

Hablame siempre en español.

## Comandos esenciales

```bash
go run ./cmd/api              # API (entrypoint: cmd/api/main.go)
go test ./... -v              # Tests
./migrate.sh up               # Migraciones (si el script existe en el repo)
make migrate                  # Alternativa: Makefile (puede cubrir subconjunto de SQL)
```

## Contexto on-demand (cargar según necesidad)

| Archivo | Cuándo cargar |
|---------|---------------|
| `tenant-service-management/api-endpoints.md` | Rutas, payloads JSON, bootstrap, locking |
| `tenant-service-management/architecture.md` | Capas hexagonales, eventos, reglas de negocio |
| `tenant-service-management/config.md` | Variables de entorno, migraciones |

## Reglas compartidas (workspace `saas-mt`, cargar según contexto)

| Regla | Archivo |
|-------|---------|
| Arquitectura hexagonal | `ai-tools/rules/architecture.md` |
| Multi-tenancy | `ai-tools/rules/multi-tenant.md` |
| API Gateway / Kong | `ai-tools/rules/api-gateway.md` |
| Formato respuesta API | `ai-tools/rules/api-response-format.md` |
| Testing | `ai-tools/rules/testing-standards.md` |
| Seguridad base | `ai-tools/rules/security-baseline.md` |

Antes de generar mucho código Go nuevo, valorar MCP **mcp-go-generator-node** según el flujo del equipo.
