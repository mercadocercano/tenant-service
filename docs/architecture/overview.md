# Arquitectura — Tenant Service

> Archivo movido desde `ARCHITECTURE.md` en la raíz. El original queda como referencia histórica.

## Visión general

El **Tenant Service** gestiona la configuración operativa y fiscal de cada tenant en el ecosistema SaaS multi-tenant Mercado Cercano. Es la única fuente de verdad para políticas de negocio (stock, crédito, fiscalidad) y puntos de venta.

## Componentes principales

| Componente | Responsabilidad | Tecnología |
|-----------|----------------|-----------|
| `domain/entity` | Entidades y reglas de negocio (TenantSettings, PointOfSale, TenantConfig) | Go puro |
| `domain/repository` | Interfaces de persistencia (ports) | Go interfaces |
| `application/command` | Write use cases con optimistic locking | Go |
| `application/query` | Read use cases | Go |
| `infrastructure/persistence` | Adaptadores PostgreSQL | `database/sql` + `lib/pq` |
| `infrastructure/controller` | HTTP handlers | Gin |
| `infrastructure/event` | Publicación de domain events | EventBus adapter |

## Capas y regla de dependencia

```
domain ← application ← infrastructure
```

El dominio no importa nada del proyecto. La aplicación importa solo interfaces de dominio. La infraestructura importa la aplicación y el dominio — nunca al revés.

## Flujo principal — GET /api/v1/tenant/settings

```
HTTP Request
  → TenantSettingsController (infrastructure/controller)
    → GetTenantSettingsQuery (application/query)
      → TenantSettingsRepository interface (domain/repository)
        → PostgresTenantSettingsRepository (infrastructure/persistence)
          → tenant_settings table
  → TenantSettingsResponse DTO (application/response)
  → HTTP 200 JSON
```

## Decisiones de arquitectura relevantes

- [ADR-001: Arquitectura Hexagonal y DDD](../adr/ADR-001-arquitectura-hexagonal-ddd.md)
- [ADR-002: Separación tenant_db y eventbus](../adr/ADR-002-separacion-tenant-db-eventbus.md)
- [ADR-003: Domain Events vía EventBus](../adr/ADR-003-domain-events-via-eventbus.md)

## Multi-tenancy

Row-level isolation: cada tabla incluye `tenant_id`. El middleware `TenantValidation` (del paquete `github.com/mercadocercano/middleware`) valida el header `X-Tenant-ID` y el token JWT en cada request.

## Modelo de datos

Tres tablas en `tenant_db`:
- `tenant_config` — key-value flexible para configuraciones extensibles
- `tenant_settings` — configuración estructurada core (fiscal, stock, crédito, moneda)
- `points_of_sale` — puntos de venta del tenant (sucursales)

`tenant_settings` usa optimistic locking via campo `version`.
