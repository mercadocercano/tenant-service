# Guía: Endpoint Bootstrap

> Archivo movido desde `BOOTSTRAP_ENDPOINT.md` en la raíz.

## Descripción

`POST /api/v1/tenant/bootstrap` inicializa la configuración por defecto de un tenant durante el onboarding.

- **Idempotente**: Se puede llamar múltiples veces sin efectos secundarios
- **Best-effort**: No bloquea el onboarding si falla
- **Uso interno**: Llamado por el `onboarding-service` al completar el proceso

## Request

```bash
curl -X POST http://localhost:8125/api/v1/tenant/bootstrap \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer <jwt>"
```

Body vacío. El tenant ID viene del header.

## Response (200 OK)

```json
{
  "success": true,
  "message": "Tenant configuration bootstrapped successfully",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "configs_created": 1
}
```

## Configuraciones inicializadas

| Key | Valor default |
|-----|---------------|
| `catalog.stock_policy` | `REQUIRE_STOCK` |

## Integración con Onboarding Service

```go
// En onboarding-service:
// TENANT_SERVICE_URL=http://mc-tenant-service:8120
uc.bootstrapTenantConfigAsync(process)  // async, best-effort
```

## Troubleshooting

- **"X-Tenant-ID header is required"** → Verificar que el header esté presente
- **Config no se crea en segunda llamada** → Correcto (idempotente, ya existe)
- **Best-effort warning** → El bootstrap falló pero el onboarding continuó; la config se puede crear manualmente o re-intentar llamando al endpoint directamente
