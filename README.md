# Tenant Service

Centro de configuración operativa y fiscal del tenant. Gestiona políticas de negocio (stock, crédito, fiscalidad), puntos de venta y configuraciones key-value.

**Puerto**: 8120 (externo lab: 8125) | **Stack**: Go + Gin + PostgreSQL + EventBus

## API Endpoints

| Método | Path | Descripción |
|--------|------|-------------|
| `GET` | `/api/v1/tenant/settings` | Configuración estructurada del tenant |
| `PUT` | `/api/v1/tenant/settings` | Actualizar settings (optimistic locking, requiere `version`) |
| `POST` | `/api/v1/tenant/points-of-sale` | Crear punto de venta |
| `GET` | `/api/v1/tenant/points-of-sale` | Listar puntos de venta (`?only_active=true`) |
| `GET` | `/api/v1/tenant/config/:key` | Obtener config key-value |
| `POST` | `/api/v1/tenant/config` | Guardar/actualizar config key-value |
| `POST` | `/api/v1/tenant/bootstrap` | Inicializar configuración por defecto (idempotente) |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Métricas Prometheus |

Headers requeridos: `X-Tenant-ID: <uuid>`, `Authorization: Bearer <jwt>`.

## Eventos publicados

| Evento | Trigger | Consumidores |
|--------|---------|-------------|
| `tenant.settings.updated` v1 | `PUT /settings` exitoso | sales, ledger, fiscal, stock |
| `tenant.point_of_sale.created` v1 | `POST /points-of-sale` exitoso | fiscal |

## Reglas de negocio

### Optimistic locking en `tenant_settings`

El cliente envía el `version` leído. Si cambió en BD → `409 Conflict`. Si no existe → insert. La versión se incrementa en cada update exitoso.

### Validaciones de `tenant_settings`

- `base_currency` debe estar en `allowed_currencies`
- `fiscal_mode`: `DISABLED` | `OPTIONAL` | `REQUIRED`
- `stock_policy`: `IGNORE` | `RESERVE` | `DEDUCT`
- `invoice_generation`: `MANUAL` | `AUTO_ON_SALE` | `AUTO_ON_CONFIRM`
- `max_credit_limit >= 0`, `default_credit_days >= 0`

### Defaults al crear un tenant

```
base_currency=ARS  fiscal_mode=DISABLED  stock_policy=IGNORE  credit_enabled=false  version=1
```

## Integración con otros servicios

Los consumidores del evento `tenant.settings.updated` usan cada campo de settings para sus políticas:

- **sales-service**: valida stock (`stock_policy`), crédito (`credit_enabled`, `max_credit_limit`), modo fiscal
- **ledger-service**: convierte montos a `base_currency`, usa `cash_customer_id`
- **fiscal-service**: genera facturas según `fiscal_mode` e `invoice_generation`
- **stock-service**: aplica `stock_policy`

Para consumir la API directamente desde otro servicio en `lab-network`:

```go
url := fmt.Sprintf("http://mc-tenant-service:8120/api/v1/tenant/config/%s", key)
req.Header.Set("X-Tenant-ID", tenantID)
```

## Documentación

Ver [`docs/`](docs/README.md) para arquitectura detallada, ADRs, setup y guías operativas.

| Documento | Descripción |
|-----------|-------------|
| [Arquitectura](docs/architecture/overview.md) | Capas, componentes, modelo de datos, multi-tenancy |
| [Write Path](docs/architecture/write-path.md) | Optimistic locking, domain events, validaciones |
| [Getting Started](docs/setup/getting-started.md) | Variables de entorno, cómo levantar, tests |
| [Deployment](docs/runbooks/deployment.md) | Deploy con Docker, migraciones, rollback |
| [Kong Integration](docs/guides/kong-integration.md) | Rutas, plugins, consumo desde otros servicios |
| [Bootstrap Endpoint](docs/guides/bootstrap-endpoint.md) | Inicialización de tenant |
| [Auditoría](docs/development/audit.md) | Reporte de arquitectura y cobertura de tests |

### ADRs

| ADR | Decisión |
|-----|----------|
| [ADR-001](docs/adr/ADR-001-arquitectura-hexagonal-ddd.md) | Arquitectura Hexagonal y DDD |
| [ADR-002](docs/adr/ADR-002-separacion-tenant-db-eventbus.md) | Separación tenant_db y eventbus |
| [ADR-003](docs/adr/ADR-003-domain-events-via-eventbus.md) | Domain Events vía EventBus |
