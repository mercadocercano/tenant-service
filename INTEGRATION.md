# Integración del Tenant Service

## 🔗 Integración con Kong Gateway

El `tenant-service` está integrado en el API Gateway de Kong y es accesible a través de:

```
http://localhost:8001/tenant/api/v1/tenant/config/:key
```

### Headers Requeridos

- `Authorization: Bearer <jwt-token>` - Token JWT del usuario autenticado
- `X-Tenant-ID: <tenant-uuid>` - UUID del tenant

### Ejemplo de Uso

```bash
# Obtener configuración de stock policy
curl -X GET http://localhost:8001/tenant/api/v1/tenant/config/catalog.stock_policy \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

## 🔌 Consumo desde otros Servicios

### Desde catalog-bff-service (Go)

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type TenantConfigResponse struct {
    Key   string  `json:"key"`
    Value *string `json:"value"`
}

func GetTenantConfig(ctx context.Context, tenantID, key string) (*string, error) {
    url := fmt.Sprintf("http://tenant-service:8120/api/v1/tenant/config/%s", key)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("X-Tenant-ID", tenantID)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 404 {
        return nil, nil // Config no existe
    }
    
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    
    var config TenantConfigResponse
    if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
        return nil, err
    }
    
    return config.Value, nil
}

// Uso con fallback
func GetStockPolicy(ctx context.Context, tenantID string) string {
    policy, err := GetTenantConfig(ctx, tenantID, "catalog.stock_policy")
    if err != nil || policy == nil {
        return "VALIDATE_STOCK" // Fallback por defecto
    }
    return *policy
}
```

### Desde Frontends (Next.js)

```typescript
// lib/tenant-config.ts
export async function getTenantConfig(key: string): Promise<string | null> {
  try {
    const response = await fetch(
      `http://localhost:8001/tenant/api/v1/tenant/config/${key}`,
      {
        headers: {
          'Authorization': `Bearer ${getToken()}`,
          'X-Tenant-ID': getTenantId(),
        },
      }
    );

    if (response.status === 404) {
      return null; // Config no existe
    }

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    return data.value;
  } catch (error) {
    console.error('Error getting tenant config:', error);
    return null;
  }
}

// Uso con fallback
export async function getStockPolicy(): Promise<string> {
  const policy = await getTenantConfig('catalog.stock_policy');
  return policy || 'VALIDATE_STOCK'; // Fallback
}
```

## ✍️ Escritura de Configuraciones

### Endpoint POST (v1.1.0+)

```bash
POST /api/v1/tenant/config
Headers:
  X-Tenant-ID: <tenant-uuid>
  Authorization: Bearer <jwt-token>
  Content-Type: application/json

Body:
{
  "key": "catalog.stock_policy",
  "value": "IGNORE_STOCK"
}
```

**Ejemplo desde Go:**
```go
func SetTenantConfig(ctx context.Context, tenantID, key, value string) error {
    url := "http://tenant-service:8120/api/v1/tenant/config"
    
    body := map[string]string{
        "key":   key,
        "value": value,
    }
    
    jsonBody, _ := json.Marshal(body)
    req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
    req.Header.Set("X-Tenant-ID", tenantID)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    
    return nil
}
```

**Ejemplo desde TypeScript:**
```typescript
async function setTenantConfig(key: string, value: string): Promise<void> {
  const response = await fetch(
    'http://localhost:8001/tenant/api/v1/tenant/config',
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getToken()}`,
        'X-Tenant-ID': getTenantId(),
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ key, value }),
    }
  );

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
}
```

### Inserción Manual (Alternativa)

También puedes insertar configuraciones directamente en la BD:

```sql
-- Conectar a la base de datos
psql -h localhost -p 5435 -U postgres -d tenant_db

-- Insertar configuración
INSERT INTO tenant_config (tenant_id, config_key, config_value)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'catalog.stock_policy', 'IGNORE_STOCK')
ON CONFLICT (tenant_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_at = now();

-- Verificar
SELECT * FROM tenant_config WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
```

## 📋 Configuraciones Disponibles (v1)

### catalog.stock_policy

Controla el comportamiento de validación de stock en el catálogo.

**Valores posibles:**
- `VALIDATE_STOCK` - Valida stock antes de publicar/vender (default)
- `IGNORE_STOCK` - No valida stock, permite productos sin stock

**Ejemplo:**
```sql
INSERT INTO tenant_config (tenant_id, config_key, config_value)
VALUES ('tenant-uuid', 'catalog.stock_policy', 'IGNORE_STOCK');
```

## 🚀 Próximas Configuraciones (Roadmap)

### catalog.auto_publish
```sql
-- Auto-publicar productos al crearlos
VALUES ('tenant-uuid', 'catalog.auto_publish', 'true');
```

### integrations.mercadopago.api_key
```sql
-- API Key de MercadoPago
VALUES ('tenant-uuid', 'integrations.mercadopago.api_key', 'APP_USR_...');
```

### templates.onboarding_type
```sql
-- Tipo de negocio para templates
VALUES ('tenant-uuid', 'templates.onboarding_type', 'almacen');
```

## 🔒 Seguridad

- **Autenticación**: JWT obligatorio (excepto health check)
- **Multi-tenancy**: Validación automática de `X-Tenant-ID`
- **Rate Limiting**: 100 req/min, 2000 req/hora
- **ACL**: Solo usuarios autenticados

## 📊 Monitoreo

### Métricas Prometheus

```
# Endpoint de métricas
http://localhost:8120/metrics
```

### Health Check

```bash
curl http://localhost:8120/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "tenant-service",
  "version": "1.0.0"
}
```

## 🐛 Troubleshooting

### Error: "X-Tenant-ID header is required"

Asegúrate de enviar el header `X-Tenant-ID` en todas las peticiones.

### Error: "tenant config not found"

La configuración no existe en la BD. Insértala manualmente o usa un fallback en el consumidor.

### Error: "connection refused"

Verifica que el servicio esté corriendo:

```bash
docker-compose ps tenant-service
docker-compose logs tenant-service
```

## 📝 Notas Importantes

1. **El servicio NO interpreta valores**: Solo almacena y devuelve strings. La semántica vive en el consumidor.

2. **Fallbacks son responsabilidad del consumidor**: Si una config no existe (404), el servicio que consume debe tener un valor por defecto.

3. **Sin cache (v1)**: Cada request va a la BD. Considera cachear en el consumidor si es crítico.

4. **Escritura disponible (v1.1.0+)**: Usa el endpoint POST para upsert de configuraciones.

## 🔗 Referencias

- [README.md](./README.md) - Documentación general del servicio
- [kong-config.yml](./kong-config.yml) - Configuración de Kong standalone
- [API Gateway Kong](../api-gateway/kong.yml) - Configuración integrada
