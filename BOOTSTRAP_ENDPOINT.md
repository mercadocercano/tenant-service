# Endpoint Bootstrap - Tenant Service

## Descripción

El endpoint `/api/v1/tenant/bootstrap` inicializa la configuración por defecto de un tenant nuevo durante el proceso de onboarding.

## Características

- **Idempotente**: Puede llamarse múltiples veces sin efectos secundarios
- **Best-effort**: No bloquea el onboarding si falla
- **Interno**: Solo accesible por servicios internos (no usuarios finales)

## Endpoint

```
POST /api/v1/tenant/bootstrap
```

### Headers Requeridos

```
X-Tenant-ID: <uuid>
Authorization: Bearer <service-jwt>  (opcional en desarrollo)
```

### Request Body

```json
{}
```

El body está vacío. El tenant ID se obtiene del header `X-Tenant-ID`.

### Response (200 OK)

```json
{
  "success": true,
  "message": "Tenant configuration bootstrapped successfully",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "configs_created": 1
}
```

### Configuraciones Default Creadas

| Key | Value | Descripción |
|-----|-------|-------------|
| `catalog.stock_policy` | `REQUIRE_STOCK` | Política de stock por defecto |

## Integración con Onboarding

El onboarding-service llama a este endpoint automáticamente al completar el proceso de onboarding:

```go
// En CompleteOnboardingUseCase.Execute()
uc.bootstrapTenantConfigAsync(process)
```

La llamada es **asíncrona** y **best-effort**:
- No bloquea la respuesta del onboarding
- Si falla, solo registra un warning en los logs
- El tenant puede funcionar sin config inicial (se creará después)

## Seguridad

### Kong Gateway ACL

El endpoint está protegido por ACL en Kong:

```yaml
- name: tenant-bootstrap-route
  paths:
    - /tenant/api/v1/tenant/bootstrap
  strip_path: false
  methods:
    - POST
  plugins:
    - name: acl
      config:
        allow:
          - "service-internal"
```

Solo servicios con el grupo ACL `service-internal` pueden acceder.

### Consumer para Onboarding Service

```yaml
- username: onboarding-service-consumer
  custom_id: onboarding-service
  acls:
    - group: "service-internal"
    - group: "authenticated"
```

## Variables de Entorno

### Onboarding Service

```bash
TENANT_SERVICE_URL=http://tenant-service:8120
```

En desarrollo local:
```bash
TENANT_SERVICE_URL=http://localhost:8120
```

## Testing

### Test Unitario

```bash
cd services/tenant-service
go test ./src/tenant/application/command/bootstrap_tenant_config_command_test.go -v
```

### Test de Integración

```bash
# 1. Iniciar servicios
make dev-start

# 2. Completar onboarding (incluye bootstrap automático)
curl -X POST http://localhost:8001/onboarding/api/v1/complete \
  -H "Content-Type: application/json" \
  -d '{
    "process_id": "550e8400-e29b-41d4-a716-446655440000"
  }'

# 3. Verificar config creada
curl -X GET http://localhost:8001/tenant/api/v1/tenant/config/catalog.stock_policy \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer <jwt-token>"
```

### Test Manual del Endpoint

```bash
# Llamada directa (solo para testing, normalmente es llamado por onboarding)
curl -X POST http://localhost:8120/api/v1/tenant/bootstrap \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Logs

### Onboarding Service

```
=== BOOTSTRAP TENANT CONFIG PROCESS START ===
Tenant client available, starting bootstrap process asynchronously...
=== ASYNC BOOTSTRAP TENANT CONFIG GOROUTINE START ===
Attempting to bootstrap tenant config for TenantID: 550e8400-e29b-41d4-a716-446655440000
SUCCESS: Tenant config bootstrapped successfully
=== ASYNC BOOTSTRAP TENANT CONFIG GOROUTINE END ===
```

### Tenant Service

```
=== BOOTSTRAP ENDPOINT START ===
TenantID from header: 550e8400-e29b-41d4-a716-446655440000
Executing bootstrap command for tenant: 550e8400-e29b-41d4-a716-446655440000
=== BOOTSTRAP TENANT CONFIG START ===
Checking config: catalog.stock_policy
Creating default config: catalog.stock_policy = REQUIRE_STOCK
Config catalog.stock_policy created successfully
Bootstrap completed: 1 configs created
=== BOOTSTRAP TENANT CONFIG END ===
Bootstrap completed successfully: 1 configs created
=== BOOTSTRAP ENDPOINT END ===
```

## Troubleshooting

### Error: "X-Tenant-ID header is required"

El header `X-Tenant-ID` es obligatorio. Verificar que el onboarding service lo esté enviando.

### Warning: "Failed to bootstrap tenant config (best-effort, continuing)"

El onboarding continuó exitosamente aunque falló el bootstrap. Posibles causas:
- Tenant service no está disponible
- Timeout de red
- Error de base de datos en tenant service

El tenant puede funcionar normalmente. La config se puede crear manualmente después.

### Configs no se crean en segunda llamada

Esto es correcto (idempotencia). El endpoint verifica si ya existen y no las duplica.

## Arquitectura

### Desacoplamiento de Dominios

```
┌─────────────────────┐
│ Onboarding Service  │
│  (Domain: Onboarding)│
└──────────┬──────────┘
           │ HTTP POST (best-effort)
           │ /api/v1/tenant/bootstrap
           ▼
┌─────────────────────┐
│  Tenant Service     │
│  (Domain: Config)   │
└─────────────────────┘
```

**Principios aplicados:**
- ✅ Onboarding NO decide policies
- ✅ Onboarding NO persiste config local
- ✅ tenant-service es el dueño de la config
- ✅ Integración best-effort (no bloquea onboarding)
- ✅ Sin imports cruzados entre dominios

## Extensión Futura

Para agregar nuevas configuraciones default:

```go
// En bootstrap_tenant_config_command.go
var DefaultConfigs = map[string]string{
	"catalog.stock_policy": "REQUIRE_STOCK",
	"orders.auto_approve": "false",           // Nueva config
	"notifications.email_enabled": "true",    // Nueva config
}
```

El comando es genérico y procesará todas las configs del mapa.
