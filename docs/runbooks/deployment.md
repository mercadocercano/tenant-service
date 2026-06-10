# Runbook: Deployment

> Archivo movido desde `DEPLOYMENT.md` en la raíz. Rutas de proyecto actualizadas.

**Frecuencia**: On-demand (al hacer deploy)
**Duración estimada**: 2-5 minutos

## Pasos

### 1. Levantar con Docker Compose (lab)

```bash
# Desde el directorio del servicio
docker compose up -d

# Ver logs
docker compose logs -f tenant-service

# Verificar salud
curl http://localhost:8125/health
```

### 2. Migraciones

Las migraciones SQL están en `migrations/`. Se aplican automáticamente vía `postgres-setup` al iniciar el servicio con docker compose. Para aplicarlas manualmente:

```bash
psql -h localhost -U postgres -d tenant_db -f migrations/001_create_tenant_config_table.sql
psql -h localhost -U postgres -d tenant_db -f migrations/003_create_tenant_settings.sql
psql -h localhost -U postgres -d tenant_db -f migrations/004_create_points_of_sale.sql
```

### 3. Verificación post-deploy

```bash
# Health check
curl http://localhost:8125/health

# Métricas
curl http://localhost:8125/metrics | grep go_info

# Via Kong
curl -sf http://localhost:8000/tenant/health && echo "Kong OK"
```

## Rollback

```bash
# Bajar el servicio
docker compose down

# Restaurar la imagen anterior (cambiar tag en docker-compose.yml)
docker compose up -d
```

## Indicadores de éxito

- `curl /health` retorna `{"status":"healthy"}`
- El container aparece en `lab-network`: `docker network inspect lab-network`
- Sin errores en `docker compose logs --tail 20`

## Notas

- El container se llama `mc-tenant-service` en la red `lab-network`
- Puerto externo: `8125`, puerto interno: `8120`
- El build de desarrollo requiere `GITHUB_TOKEN` para módulos privados; el build de producción los descarga via cache o registry
