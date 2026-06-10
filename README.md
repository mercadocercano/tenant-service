# Tenant Service

Microservicio de configuración centralizada para el sistema SaaS Multi-Tenant.

## Descripción

El Tenant Service es el **centro de configuración operativa y fiscal** del sistema. Proporciona:

- **Configuraciones estructuradas** (tenant_settings): Políticas fiscales, monetarias, de stock y crédito
- **Puntos de venta** (points_of_sale): Gestión de sucursales y puntos fiscales
- **Configuraciones flexibles** (tenant_config): Key-value para configs experimentales

## Arquitectura

### Modelo Híbrido

El servicio utiliza un **modelo híbrido** que combina:

1. **Configuraciones Core Estructuradas** (`tenant_settings`):
   - Tipado fuerte
   - Validaciones automáticas
   - Performance optimizada (1 query)
   - Optimistic locking con versioning
   - Publicación de eventos cuando cambian

2. **Configuraciones Flexibles** (`tenant_config`):
   - Key-value genérico
   - Para configs experimentales
   - Para features temporales
   - Para personalizaciones custom

### Configuraciones Estructuradas

La tabla `tenant_settings` centraliza:

#### Monetaria
- `base_currency`: Moneda base del tenant (ej: ARS, USD)
- `allowed_currencies`: Monedas permitidas
- `exchange_rate_source`: MANUAL | EXTERNAL_API
- `auto_update_exchange_rate`: Actualización automática

#### Fiscal
- `fiscal_mode`: DISABLED | OPTIONAL | REQUIRED
- `invoice_generation`: MANUAL | AUTO_ON_SALE | AUTO_ON_CONFIRM
- `default_invoice_type`: A, B, C, etc.
- `tax_regime`: MONOTRIBUTO | RESPONSABLE_INSCRIPTO
- `allow_sale_if_afip_fails`: Permitir venta si falla AFIP
- `auto_retry_failed_invoices`: Reintentar facturas fallidas
- `email_invoice_after_success`: Enviar email tras facturar

#### Stock
- `stock_policy`: IGNORE | RESERVE | DEDUCT
- `allow_negative_stock`: Permitir stock negativo
- `require_stock_validation_before_sale`: Validar antes de vender

#### Crédito
- `credit_enabled`: Habilitar crédito
- `default_credit_days`: Días de crédito por defecto
- `max_credit_limit`: Límite máximo de crédito
- `allow_sale_over_credit_limit`: Permitir venta sobre límite

#### Cliente Contado
- `cash_customer_id`: UUID del cliente genérico para ventas al contado

## API Endpoints

### Configuraciones Estructuradas

#### `GET /api/v1/tenant/settings`
Obtiene todas las configuraciones del tenant.

**Headers**:
- `X-Tenant-ID`: UUID del tenant

**Response**: `200 OK`
```json
{
  "tenant_id": "uuid",
  "base_currency": "ARS",
  "allowed_currencies": ["ARS", "USD"],
  "fiscal_mode": "REQUIRED",
  "stock_policy": "RESERVE",
  "credit_enabled": true,
  "max_credit_limit": 500000.00,
  "version": 2,
  "updated_at": "2026-02-19T10:00:00Z"
}
```

#### `PUT /api/v1/tenant/settings`
Actualiza configuraciones con optimistic locking.

**Headers**:
- `X-Tenant-ID`: UUID del tenant

**Body**:
```json
{
  "version": 2,
  "base_currency": "ARS",
  "allowed_currencies": ["ARS", "USD"],
  "fiscal_mode": "REQUIRED",
  "invoice_generation": "AUTO_ON_SALE",
  "stock_policy": "RESERVE",
  "credit_enabled": true,
  "default_credit_days": 30,
  "max_credit_limit": 500000.00,
  "cash_customer_id": "uuid",
  ...
}
```

**Response**: 
- `200 OK`: Actualizado correctamente
- `409 Conflict`: Version conflict (otro proceso modificó la configuración)

### Puntos de Venta

#### `POST /api/v1/tenant/points-of-sale`
Crea un nuevo punto de venta.

**Body**:
```json
{
  "code": 1,
  "description": "Sucursal Centro",
  "is_fiscal_enabled": true,
  "default_invoice_type": "B"
}
```

#### `GET /api/v1/tenant/points-of-sale?only_active=true`
Lista puntos de venta del tenant.

**Response**: `200 OK`
```json
{
  "points_of_sale": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "code": 1,
      "description": "Sucursal Centro",
      "is_fiscal_enabled": true,
      "default_invoice_type": "B",
      "is_active": true,
      "version": 1
    }
  ],
  "total_count": 1
}
```

### Configuraciones Key-Value (Legacy)

#### `GET /api/v1/tenant/config/:key`
Obtiene una configuración específica.

#### `POST /api/v1/tenant/config`
Guarda/actualiza una configuración.

#### `POST /api/v1/tenant/bootstrap`
Inicializa configuraciones por defecto.

## Eventos Publicados

### `tenant.settings.updated` v1

Publicado cuando se actualiza la configuración del tenant.

**Payload**:
```json
{
  "event_type": "tenant.settings.updated",
  "event_version": 1,
  "aggregate_type": "tenant_settings",
  "aggregate_id": "tenant-uuid",
  "tenant_id": "tenant-uuid",
  "occurred_at": "2026-02-19T10:00:00Z",
  "payload": {
    "version": 3,
    "base_currency": "ARS",
    "fiscal_mode": "REQUIRED",
    "stock_policy": "RESERVE",
    "credit_enabled": true,
    "max_credit_limit": 500000.00
  }
}
```

**Consumidores**:
- `sales-service`: Para validar stock antes de venta
- `ledger-service`: Para conversión de monedas
- `fiscal-service`: Para generación automática de facturas
- `stock-service`: Para aplicar política de stock

## Desarrollo

### Prerrequisitos

- Go 1.22+
- PostgreSQL 15+
- EventBus activo

### Variables de Entorno

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=tenant_db

# EventBus
EVENTBUS_DB_HOST=localhost
EVENTBUS_DB_PORT=5432
EVENTBUS_DB_USER=postgres
EVENTBUS_DB_PASSWORD=postgres
EVENTBUS_DB_NAME=eventbus

# Service
PORT=8120
GIN_MODE=debug
PROMETHEUS_ENABLED=true
```

### Ejecutar Tests

```bash
cd services/tenant-service
go test ./... -v
```

### Ejecutar Migraciones

```bash
# Dentro del container
./migrate.sh up

# Localmente
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/tenant_db?sslmode=disable" up
```

## Migraciones

### 001_create_tenant_config_table.sql
Tabla key-value para configuraciones flexibles.

### 002_seed_initial_data.sql
Datos iniciales (comentado por defecto).

### 003_create_tenant_settings.sql
Tabla estructurada para configuraciones core.

### 004_create_points_of_sale.sql
Tabla de puntos de venta por tenant.

## Arquitectura Hexagonal

```
src/tenant/
├── domain/
│   ├── entity/           # TenantSettings, PointOfSale
│   ├── repository/       # Interfaces
│   └── valueobject/      # ConfigKey
├── application/
│   ├── command/          # UpdateTenantSettings, CreatePointOfSale
│   ├── query/            # GetTenantSettings, ListPointsOfSale
│   ├── request/          # DTOs de entrada
│   └── response/         # DTOs de salida
└── infrastructure/
    ├── controller/       # HTTP handlers
    ├── persistence/      # PostgreSQL repositories
    ├── event/            # EventPublisher adapter
    └── config/           # Wiring e inyección de dependencias
```

## Reglas de Negocio

### Optimistic Locking

Todas las actualizaciones de `tenant_settings` usan optimistic locking:

1. Cliente lee configuración (obtiene `version`)
2. Cliente modifica y envía con `version` original
3. Servicio valida que `version` en DB sea la misma
4. Si coincide: actualiza e incrementa `version`
5. Si no coincide: retorna `409 Conflict`

### Validaciones

- `base_currency` debe estar en `allowed_currencies`
- `max_credit_limit` >= 0
- `default_credit_days` >= 0
- Valores enumerados deben ser válidos (ej: fiscal_mode)

### Valores por Defecto

Al crear un tenant nuevo:
- `base_currency`: "ARS"
- `fiscal_mode`: "DISABLED"
- `stock_policy`: "IGNORE"
- `credit_enabled`: false
- `version`: 1

## Integración con Otros Servicios

### Sales Service
Consume `tenant.settings.updated` para:
- Validar stock según `stock_policy`
- Aplicar `fiscal_mode` al crear ventas
- Validar crédito según `credit_enabled` y `max_credit_limit`

### Ledger Service
Consume `tenant.settings.updated` para:
- Convertir montos a `base_currency`
- Usar `cash_customer_id` para ventas al contado
- Validar límites de crédito

### Fiscal Service
Consume `tenant.settings.updated` para:
- Generar facturas según `fiscal_mode`
- Aplicar `invoice_generation` automática
- Usar `default_invoice_type`

## Documentación

Ver [`docs/`](docs/README.md) para arquitectura, ADRs, setup y guías operativas.

| Documento | Descripción |
|-----------|-------------|
| [Arquitectura](docs/architecture/overview.md) | Capas, componentes, modelo de datos |
| [Write Path](docs/architecture/write-path.md) | Optimistic locking, domain events |
| [Getting Started](docs/setup/getting-started.md) | Variables de entorno, cómo levantar |
| [Deployment](docs/runbooks/deployment.md) | Deploy, migraciones, rollback |
| [Kong Integration](docs/guides/kong-integration.md) | Rutas, plugins, consumo desde otros servicios |
| [Bootstrap Endpoint](docs/guides/bootstrap-endpoint.md) | Inicialización de tenant |

### ADRs

| ADR | Decisión |
|-----|----------|
| [ADR-001](docs/adr/ADR-001-arquitectura-hexagonal-ddd.md) | Arquitectura Hexagonal y DDD |
| [ADR-002](docs/adr/ADR-002-separacion-tenant-db-eventbus.md) | Separación tenant_db y eventbus |
| [ADR-003](docs/adr/ADR-003-domain-events-via-eventbus.md) | Domain Events vía EventBus |

## Contacto

Para consultas sobre este servicio, contactar al equipo de arquitectura.
