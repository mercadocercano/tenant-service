# Write Path — Tenant Service

> Archivo movido desde `WRITE_PATH_IMPLEMENTATION.md` en la raíz.

**Versión**: 1.1.0 | **Estado**: Completado

## Endpoints de escritura

| Método | Path | Descripción |
|--------|------|-------------|
| POST | `/api/v1/tenant/config` | Upsert de config key-value |
| POST | `/api/v1/tenant/bootstrap` | Bootstrap de configuración inicial (idempotente) |
| PUT | `/api/v1/tenant/settings` | Actualización de settings con optimistic locking |
| POST | `/api/v1/tenant/points-of-sale` | Crear punto de venta |

## Optimistic locking en tenant_settings

El `UpdateTenantSettingsCommand` recibe el `version` actual del cliente. El repositorio:
1. Intenta `UPDATE ... WHERE tenant_id = $1 AND version = $currentVersion`
2. Si `rowsAffected == 0`:
   - Verifica si existe → si existe, retorna error "version conflict"
   - Si no existe → hace `INSERT`

## Publicación de eventos

Al completar exitosamente un command de escritura, se publican eventos de dominio:
- `UpdateTenantSettingsCommand` → `tenant.settings.updated`
- `CreatePointOfSaleCommand` → `tenant.point_of_sale.created`

La publicación es best-effort (fallo no revierte la escritura). Ver [ADR-003](../adr/ADR-003-domain-events-via-eventbus.md).

## Validaciones

### TenantSettings
- `fiscal_mode`: DISABLED | OPTIONAL | REQUIRED
- `invoice_generation`: MANUAL | AUTO_ON_SALE | AUTO_ON_CONFIRM
- `stock_policy`: IGNORE | RESERVE | DEDUCT
- `tax_regime`: MONOTRIBUTO | RESPONSABLE_INSCRIPTO
- `exchange_rate_source`: MANUAL | EXTERNAL_API
- `base_currency` debe estar en `allowed_currencies`
- `max_credit_limit >= 0`, `default_credit_days >= 0`

### PointOfSale
- `code > 0`
- `description` no vacío
- `default_invoice_type` no vacío
