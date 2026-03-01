# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [1.1.0] - 2026-02-03

### ✨ Agregado

#### Write Path - Persistencia de Configuraciones
- Endpoint POST `/api/v1/tenant/config` para escritura de configuraciones
- Command `SetTenantConfigCommand` en capa de aplicación
- DTO `SetTenantConfigRequest` con validaciones
- Comportamiento upsert (INSERT si no existe, UPDATE si existe)
- Tests unitarios del command con mocks
- Tests del repositorio (happy path) con sqlmock

#### Testing
- 4 tests unitarios para `SetTenantConfigCommand`
  - Insert de nueva configuración
  - Update de configuración existente
  - Manejo de errores en GetByKey
  - Manejo de errores en Save
- 4 tests para `PostgresTenantConfigRepository`
  - Save (insert)
  - Save (update/upsert)
  - GetByKey (found)
  - GetByKey (not found)

#### Documentación
- Actualizado README.md con endpoint POST
- Ejemplos de uso con curl
- Documentación de comportamiento upsert

### 🔧 Modificado
- Controller actualizado para incluir `SetTenantConfigCommand`
- Wiring/DI actualizado en `TenantModuleConfig`
- Estructura de directorios actualizada (agregado `application/command/` y `application/request/`)

## [1.0.0] - 2026-02-03

### ✨ Agregado

#### Core Functionality
- Microservicio `tenant-service` con arquitectura hexagonal completa
- Endpoint GET `/api/v1/tenant/config/:key` para lectura de configuraciones
- Soporte multi-tenant con validación de `X-Tenant-ID` header
- Health check endpoint `/health`

#### Dominio
- Entidad `TenantConfig` como agregado raíz
- Value Object `ConfigKey` para claves namespaced
- Repository pattern con interfaz `TenantConfigRepository`
- Excepción de dominio `TenantConfigNotFound`

#### Infraestructura
- Implementación PostgreSQL del repositorio
- Controlador HTTP con Gin framework
- Migración SQL para tabla `tenant_config`
- Índices optimizados para queries por tenant y key

#### Docker & Deployment
- Dockerfile multi-stage (development, production)
- docker-compose.yml con servicio + base de datos
- Hot reload con Air para desarrollo
- Health checks configurados

#### Integración
- Configuración completa en Kong Gateway
- Plugins JWT, ACL, Rate Limiting
- CORS configurado
- Rutas registradas en `/tenant/*`

#### Testing
- Tests unitarios para entidad `TenantConfig`
- Tests unitarios para value object `ConfigKey`
- Cobertura de casos edge (validaciones, errores)

#### Documentación
- README.md completo con quick start
- ARCHITECTURE.md con decisiones de diseño
- INTEGRATION.md con ejemplos de consumo
- DEPLOYMENT.md con guía de despliegue
- CHANGELOG.md (este archivo)
- Comentarios inline en código crítico

#### Tooling
- Makefile con comandos útiles
- Script de inicialización de BD (`init-db.sh`)
- .gitignore configurado para Go
- .air.toml para hot reload

### 🔒 Seguridad

- Autenticación JWT requerida (vía Kong)
- Validación de tenant ID en cada request
- Rate limiting: 100 req/min, 2000 req/hora
- ACL para control de acceso
- Non-root user en Docker

### 📊 Métricas

- Endpoint `/metrics` para Prometheus (opcional)
- Métricas de latencia, throughput, errores

### 🎯 Configuraciones Soportadas (v1)

- `catalog.stock_policy` - Política de validación de stock
  - Valores: `VALIDATE_STOCK`, `IGNORE_STOCK`

### 🚀 Rendimiento

- Conexión a BD con pool de conexiones
- Índices optimizados para queries frecuentes
- Timeouts configurados (60s connect, read, write)

### 📝 Notas de Versión

Esta es la primera versión estable del tenant-service. El scope es **mínimo e intencional**:

- **Solo lectura**: Escritura se hace directo a BD por ahora
- **Sin cache**: Cada request va a la BD (simplicidad > performance en v1)
- **Strings genéricos**: El servicio no interpreta valores, solo los almacena

### 🔮 Próximos Pasos (v1.1)

- [ ] Endpoint POST/PUT para escritura de configuraciones
- [ ] Validaciones de schema por namespace
- [ ] Audit log de cambios
- [ ] Cache con Redis
- [ ] UI de administración en backoffice

### 🐛 Problemas Conocidos

Ninguno reportado en v1.0.0

### 🙏 Agradecimientos

- Equipo SaaS MT - Tienda Vecina
- Inspiración en Clean Architecture (Robert C. Martin)
- Hexagonal Architecture (Alistair Cockburn)
- Domain-Driven Design (Eric Evans)

---

## [Unreleased]

### En Desarrollo

- Endpoint de escritura (POST/PUT)
- Validaciones avanzadas
- Cache layer

---

**Formato de versiones**: MAJOR.MINOR.PATCH

- **MAJOR**: Cambios incompatibles en la API
- **MINOR**: Nueva funcionalidad compatible hacia atrás
- **PATCH**: Bug fixes compatibles hacia atrás

**Convención de commits**:
- `feat:` - Nueva funcionalidad
- `fix:` - Bug fix
- `docs:` - Cambios en documentación
- `refactor:` - Refactorización sin cambios funcionales
- `test:` - Agregar o modificar tests
- `chore:` - Cambios en build, CI, etc.
