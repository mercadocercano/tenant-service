# Guía: Integración con Kong Gateway

## Ruta registrada

El tenant-service está registrado en Kong con la siguiente configuración (ver `~/Projects/infra/api-gateway/kong.yml.template`):

```
URL base: http://mc-tenant-service:8120
Path:     /tenant
strip_path: false
```

**Ejemplo de acceso via Kong:**
```bash
curl http://localhost:8000/tenant/api/v1/tenant/settings \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer <jwt>"
```

## Headers requeridos

- `Authorization: Bearer <jwt>` — JWT del usuario autenticado (validado por Kong)
- `X-Tenant-ID: <uuid>` — UUID del tenant

## Plugins activos

| Plugin | Configuración |
|--------|---------------|
| `jwt` | Valida tokens JWT antes de llegar al servicio |
| `rate-limiting` | 100 req/min por IP |
| `cors` | Origins `*`, headers CORS estándar |

## Consumo desde otros servicios Go (dentro de lab-network)

```go
type TenantConfigResponse struct {
    Key   string  `json:"key"`
    Value *string `json:"value"`
}

func GetTenantConfig(ctx context.Context, tenantID, key string) (*string, error) {
    url := fmt.Sprintf("http://mc-tenant-service:8120/api/v1/tenant/config/%s", key)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("X-Tenant-ID", tenantID)

    resp, err := http.DefaultClient.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()

    if resp.StatusCode == 404 { return nil, nil }

    var config TenantConfigResponse
    json.NewDecoder(resp.Body).Decode(&config)
    return config.Value, nil
}
```

## Aplicar cambios en Kong

Después de editar `~/Projects/infra/api-gateway/kong.yml.template`, rebuild + recrear `lab-kong`
(el template se copia a la imagen en el build, no está montado como volumen — un `restart` solo
reusa la imagen vieja):

```bash
docker compose -f ~/Projects/infra/docker-compose.yml up -d --build kong
```
