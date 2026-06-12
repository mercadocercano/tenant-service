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

## Memoria persistente (Engram)

Tenés acceso a memoria persistente entre sesiones vía las herramientas MCP de Engram (`mem_save`, `mem_search`, `mem_context`, etc.). Proyecto: **`mercado-cercano`** (memoria unificada del ecosistema, compartida con los demás servicios).

**Cuándo guardar** — sin esperar que te lo pidan:
- Al resolver un bug no trivial: síntoma, causa raíz, fix aplicado.
- Al tomar una decisión de diseño: qué se decidió y por qué.
- Al descubrir un patrón o convención del proyecto que no está documentada.
- Al completar una feature o refactor significativo: qué cambió y dónde.

**Cuándo buscar** — antes de empezar cualquier tarea:
- `mem_context` al inicio de sesión o tras una compaction para recuperar el estado anterior.
- `mem_search` cuando el usuario menciona algo que puede tener historial ("el bug de autenticación", "la migración de la semana pasada").

**Al cerrar sesión**: llamar `mem_session_summary` para dejar un resumen recuperable.
