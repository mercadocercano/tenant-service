# Guía de Despliegue - Tenant Service

## 🚀 Quick Start

### Opción 1: Docker Compose (Recomendado)

```bash
# Desde el directorio tenant-service
cd /Users/hornosg/MyProjects/saas-mt/services/tenant-service

# Iniciar servicio + base de datos
docker-compose up -d

# Ver logs
docker-compose logs -f tenant-service

# Verificar salud
curl http://localhost:8120/health
```

### Opción 2: Desarrollo Local (sin Docker)

```bash
# 1. Iniciar PostgreSQL (si no está corriendo)
# Puedes usar el docker-compose solo para la BD:
docker-compose up -d tenant-db

# 2. Configurar variables de entorno
export DB_HOST=localhost
export DB_PORT=5435
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=tenant_db
export PORT=8120

# 3. Ejecutar migraciones
./scripts/init-db.sh

# 4. Instalar dependencias
go mod download

# 5. Ejecutar servicio
go run cmd/api/main.go

# O con hot reload (requiere Air instalado)
air
```

## 🗄️ Configuración de Base de Datos

### Inicialización Automática

Si usas `docker-compose up`, las migraciones se ejecutan automáticamente.

### Inicialización Manual

```bash
# Ejecutar script de inicialización
./scripts/init-db.sh

# O manualmente con psql
psql -h localhost -p 5435 -U postgres -d tenant_db \
  -f migrations/001_create_tenant_config_table.sql
```

### Insertar Datos de Prueba

```sql
-- Conectar a la BD
psql -h localhost -p 5435 -U postgres -d tenant_db

-- Insertar configuración de ejemplo
INSERT INTO tenant_config (tenant_id, config_key, config_value)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'catalog.stock_policy', 'IGNORE_STOCK')
ON CONFLICT (tenant_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_at = now();

-- Verificar
SELECT * FROM tenant_config;
```

## 🔗 Integración con el Monorepo

### 1. Agregar al docker-compose principal

Edita `/Users/hornosg/MyProjects/saas-mt/docker-compose.yml`:

```yaml
services:
  # ... otros servicios ...

  tenant-service:
    build:
      context: ./services/tenant-service
      dockerfile: Dockerfile
      target: development
    container_name: tenant-service
    ports:
      - "8120:8120"
    environment:
      - DB_HOST=tenant-db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=tenant_db
      - PORT=8120
      - PROMETHEUS_ENABLED=true
    volumes:
      - ./services/tenant-service:/app
      - /app/tmp
    depends_on:
      tenant-db:
        condition: service_healthy
    networks:
      - saas-network

  tenant-db:
    image: postgres:15-alpine
    container_name: tenant-db
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=tenant_db
    ports:
      - "5435:5432"
    volumes:
      - tenant-db-data:/var/lib/postgresql/data
      - ./services/tenant-service/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - saas-network

volumes:
  tenant-db-data:
```

### 2. Agregar al Makefile principal

Edita `/Users/hornosg/MyProjects/saas-mt/Makefile`:

```makefile
# Agregar tenant-service a los comandos existentes

dev-start: ## Iniciar todos los servicios backend
	docker-compose up -d iam-service pim-service stock-service \
	  catalog-bff-service order-service tenant-service

tenant-service-logs: ## Ver logs del tenant-service
	docker-compose logs -f tenant-service

tenant-service-restart: ## Reiniciar tenant-service
	docker-compose restart tenant-service
```

### 3. Kong Gateway ya está configurado ✅

El archivo `/Users/hornosg/MyProjects/saas-mt/services/api-gateway/kong.yml` ya incluye la configuración del tenant-service.

## 🧪 Verificación del Despliegue

### 1. Health Check

```bash
curl http://localhost:8120/health
```

**Respuesta esperada:**
```json
{
  "status": "healthy",
  "service": "tenant-service",
  "version": "1.0.0"
}
```

### 2. Test de Endpoint (directo)

```bash
# Insertar dato de prueba primero
psql -h localhost -p 5435 -U postgres -d tenant_db -c \
  "INSERT INTO tenant_config (tenant_id, config_key, config_value) 
   VALUES ('00000000-0000-0000-0000-000000000001', 'catalog.stock_policy', 'IGNORE_STOCK') 
   ON CONFLICT DO NOTHING;"

# Probar endpoint
curl -X GET http://localhost:8120/api/v1/tenant/config/catalog.stock_policy \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001"
```

**Respuesta esperada:**
```json
{
  "key": "catalog.stock_policy",
  "value": "IGNORE_STOCK"
}
```

### 3. Test via Kong Gateway

```bash
# Asegúrate de que Kong esté corriendo
# Luego prueba el endpoint a través del gateway

curl -X GET http://localhost:8001/tenant/api/v1/tenant/config/catalog.stock_policy \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Authorization: Bearer <tu-jwt-token>"
```

### 4. Ejecutar Tests

```bash
cd /Users/hornosg/MyProjects/saas-mt/services/tenant-service

# Tests unitarios
go test ./...

# Tests con coverage
go test -cover ./...
```

## 📊 Monitoreo

### Métricas Prometheus

```bash
# Endpoint de métricas (si PROMETHEUS_ENABLED=true)
curl http://localhost:8120/metrics
```

### Logs

```bash
# Ver logs en tiempo real
docker-compose logs -f tenant-service

# Ver últimas 100 líneas
docker-compose logs --tail=100 tenant-service
```

### Base de Datos

```bash
# Conectar a la BD
psql -h localhost -p 5435 -U postgres -d tenant_db

# Ver todas las configuraciones
SELECT * FROM tenant_config ORDER BY tenant_id, config_key;

# Ver configuraciones de un tenant específico
SELECT * FROM tenant_config 
WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
```

## 🐛 Troubleshooting

### Servicio no inicia

```bash
# Ver logs de error
docker-compose logs tenant-service

# Verificar que la BD esté disponible
docker-compose ps tenant-db

# Reiniciar servicio
docker-compose restart tenant-service
```

### Error de conexión a BD

```bash
# Verificar que tenant-db esté healthy
docker-compose ps

# Verificar conectividad
docker-compose exec tenant-service ping tenant-db

# Verificar variables de entorno
docker-compose exec tenant-service env | grep DB_
```

### Tests fallan

```bash
# Limpiar y reconstruir
go clean -testcache
go test ./... -v
```

### Puerto 8120 ya en uso

```bash
# Encontrar proceso usando el puerto
lsof -i :8120

# Matar proceso (si es necesario)
kill -9 <PID>

# O cambiar puerto en docker-compose.yml
ports:
  - "8121:8120"  # Usar 8121 externamente
```

## 🔒 Seguridad en Producción

### Variables de Entorno

```bash
# NO uses valores por defecto en producción
export DB_PASSWORD=$(openssl rand -base64 32)
export JWT_SECRET=$(openssl rand -base64 64)
```

### Configuración de BD

```sql
-- Crear usuario específico para el servicio
CREATE USER tenant_service WITH PASSWORD 'secure-password';
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_config TO tenant_service;
```

### Kong Gateway

- Habilitar rate limiting más estricto
- Configurar SSL/TLS
- Usar JWT con rotación de claves

## 📝 Checklist de Despliegue

- [ ] Base de datos creada y migraciones ejecutadas
- [ ] Variables de entorno configuradas
- [ ] Servicio inicia correctamente (health check OK)
- [ ] Endpoint responde correctamente (test manual)
- [ ] Kong Gateway configurado y funcionando
- [ ] Logs visibles y sin errores críticos
- [ ] Métricas Prometheus disponibles (si aplica)
- [ ] Tests unitarios pasan
- [ ] Documentación actualizada

## 🚢 Despliegue a Producción

### Build de Imagen

```bash
# Build de imagen de producción
docker build --target production -t tenant-service:1.0.0 .

# Tag para registry
docker tag tenant-service:1.0.0 registry.example.com/tenant-service:1.0.0

# Push a registry
docker push registry.example.com/tenant-service:1.0.0
```

### Kubernetes (futuro)

```yaml
# deployment.yaml (ejemplo)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tenant-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: tenant-service
  template:
    metadata:
      labels:
        app: tenant-service
    spec:
      containers:
      - name: tenant-service
        image: registry.example.com/tenant-service:1.0.0
        ports:
        - containerPort: 8120
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: tenant-db-secret
              key: host
        # ... más configuración
```

## 📞 Soporte

Para problemas o preguntas:

1. Revisa la documentación: [README.md](./README.md), [ARCHITECTURE.md](./ARCHITECTURE.md)
2. Verifica los logs: `docker-compose logs tenant-service`
3. Consulta [INTEGRATION.md](./INTEGRATION.md) para ejemplos de uso
4. Contacta al equipo de SaaS MT

---

**Última actualización**: 2026-02-03  
**Versión**: 1.0.0
