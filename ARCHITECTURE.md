# Arquitectura del Tenant Service

## 🏛️ Visión General

El **Tenant Service** es el microservicio responsable de gestionar las configuraciones específicas de cada tenant en el sistema SaaS Multi-Tenant "Tienda Vecina". Actúa como el **"cerebro del tenant"**, siendo la única fuente de verdad (Source of Truth) para todas las configuraciones relacionadas con políticas de negocio, integraciones, y preferencias.

## 🎯 Principios de Diseño

### 1. Arquitectura Hexagonal (Ports & Adapters)

El servicio sigue estrictamente la arquitectura hexagonal, separando claramente:

- **Dominio**: Lógica de negocio pura, sin dependencias externas
- **Aplicación**: Casos de uso y orquestación
- **Infraestructura**: Implementaciones concretas (BD, HTTP, etc.)

### 2. Domain-Driven Design (DDD)

- **Agregado Raíz**: `TenantConfig` es el agregado raíz del módulo
- **Value Objects**: `ConfigKey` encapsula la lógica de claves namespaced
- **Repository Pattern**: Abstracción de persistencia en el dominio
- **Excepciones de Dominio**: Errores específicos del negocio

### 3. Principios SOLID

- **Single Responsibility**: Cada clase tiene una única razón para cambiar
- **Open/Closed**: Extensible sin modificar código existente
- **Liskov Substitution**: Interfaces bien definidas
- **Interface Segregation**: Contratos mínimos y específicos
- **Dependency Inversion**: Dependencias apuntan hacia abstracciones

## 📦 Estructura de Capas

```
src/
├── tenant/                      # Módulo de dominio Tenant
│   ├── domain/                  # Capa de Dominio (Core)
│   │   ├── entity/
│   │   │   └── tenant_config.go           # Agregado raíz
│   │   ├── repository/
│   │   │   └── tenant_config_repository.go # Port (interfaz)
│   │   ├── valueobject/
│   │   │   └── config_key.go              # Value Object
│   │   └── exception/
│   │       └── tenant_config_not_found.go # Excepción de dominio
│   │
│   ├── application/             # Capa de Aplicación (Use Cases)
│   │   ├── query/
│   │   │   └── get_tenant_config_query.go # Query handler
│   │   ├── request/             # DTOs de entrada (futuro)
│   │   └── response/
│   │       └── tenant_config_response.go  # DTOs de salida
│   │
│   └── infrastructure/          # Capa de Infraestructura (Adapters)
│       ├── persistence/
│       │   └── postgres_tenant_config_repository.go # Adapter de BD
│       ├── controller/
│       │   └── tenant_config_controller.go # Adapter HTTP
│       └── config/
│           └── tenant_config.go           # Wiring / DI
│
├── shared/                      # Código compartido
│   └── infrastructure/
│       └── config/
│           └── setup.go         # Configuración de middlewares
│
└── cmd/
    └── api/
        └── main.go              # Punto de entrada
```

## 🔄 Flujo de Datos

### Request Flow (GET /api/v1/tenant/config/:key)

```
1. HTTP Request
   ↓
2. Gin Router (infrastructure)
   ↓
3. TenantConfigController (infrastructure/controller)
   ├─ Valida headers (X-Tenant-ID)
   ├─ Parsea parámetros
   └─ Llama al Query Handler
   ↓
4. GetTenantConfigQuery (application/query)
   ├─ Orquesta el caso de uso
   └─ Llama al Repository
   ↓
5. TenantConfigRepository (domain/repository - interfaz)
   ↓
6. PostgresTenantConfigRepository (infrastructure/persistence)
   ├─ Ejecuta SQL
   └─ Mapea a entidad de dominio
   ↓
7. TenantConfig (domain/entity)
   ↓
8. Response DTO (application/response)
   ↓
9. HTTP Response (JSON)
```

### Dependency Direction

```
Infrastructure → Application → Domain
     ↑              ↑
     └──────────────┘
   (Solo conoce interfaces)
```

## 🗄️ Modelo de Datos

### Entidad: TenantConfig

```go
type TenantConfig struct {
    ID        uuid.UUID  // Identificador único
    TenantID  uuid.UUID  // FK al tenant (iam-service)
    Key       string     // Clave namespaced (ej: catalog.stock_policy)
    Value     string     // Valor (string genérico)
    CreatedAt time.Time  // Timestamp de creación
    UpdatedAt time.Time  // Timestamp de última actualización
}
```

### Value Object: ConfigKey

```go
type ConfigKey struct {
    value string // Formato: namespace.key
}

// Métodos:
// - Namespace() string  → "catalog"
// - Key() string        → "stock_policy"
// - String() string     → "catalog.stock_policy"
```

### Tabla: tenant_config

```sql
CREATE TABLE tenant_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    config_key VARCHAR(100) NOT NULL,
    config_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uq_tenant_config UNIQUE (tenant_id, config_key)
);

-- Índices
CREATE INDEX idx_tenant_config_tenant ON tenant_config (tenant_id);
CREATE INDEX idx_tenant_config_key ON tenant_config (config_key);
```

## 🔌 Puertos e Interfaces

### Puertos Primarios (Driving Ports)

**HTTP API** - Interfaz REST para consumidores externos

```go
// Controller
GET /api/v1/tenant/config/:key
Headers: X-Tenant-ID, Authorization
```

### Puertos Secundarios (Driven Ports)

**Repository** - Abstracción de persistencia

```go
type TenantConfigRepository interface {
    GetByKey(ctx, tenantID, key) (*TenantConfig, bool, error)
    Save(ctx, config) error
    Delete(ctx, tenantID, key) error
    GetAllByTenant(ctx, tenantID) ([]*TenantConfig, error)
}
```

## 🔐 Multi-Tenancy

### Estrategia: Row-Level Isolation

- Cada registro incluye `tenant_id`
- Validación automática en el controller
- Sin middleware específico (por ahora)
- Queries siempre filtran por `tenant_id`

### Headers Obligatorios

```
X-Tenant-ID: <uuid>     # Identifica el tenant
Authorization: Bearer   # JWT del usuario (validado por Kong)
```

## 🚀 Despliegue

### Contenedores Docker

```yaml
tenant-service:
  - Puerto: 8120
  - Base de datos: tenant-db (PostgreSQL 15)
  - Hot reload: Air (desarrollo)
  - Health check: /health
```

### Integración con Kong Gateway

```yaml
Service: tenant-service
URL: http://tenant-service:8120
Route: /tenant/*
Plugins:
  - JWT (autenticación)
  - ACL (control de acceso)
  - Rate Limiting (100/min)
  - CORS
```

## 📊 Decisiones Arquitectónicas

### ¿Por qué NO usar un config-service genérico?

- **Dominio específico**: Las configuraciones de tenant tienen semántica propia
- **Evolución independiente**: Puede crecer sin afectar otros servicios
- **Ownership claro**: Un equipo responsable de las configs de tenant

### ¿Por qué strings genéricos en lugar de tipos fuertes?

- **Flexibilidad**: Permite agregar nuevas configs sin cambiar código
- **Desacoplamiento**: El servicio no conoce la semántica de los valores
- **Responsabilidad del consumidor**: Cada servicio interpreta sus configs

### ¿Por qué NO cache en v1?

- **Simplicidad**: Empezar con lo mínimo funcional
- **Consistencia**: Siempre datos frescos de la BD
- **Futuro**: Se agregará Redis cuando sea necesario

### ¿Por qué NO endpoint de escritura en v1?

- **Scope mínimo**: Enfocarse en lectura primero
- **Seguridad**: Escritura requiere más validaciones y permisos
- **Admin manual**: Por ahora, inserciones directas a BD son suficientes

## 🔮 Roadmap Arquitectónico

### v1.0 (Actual) ✅
- Lectura de configuraciones
- Multi-tenancy básico
- Integración con Kong

### v1.1 (Próximo)
- Endpoint POST/PUT para escritura
- Validaciones de schema por namespace
- Audit log de cambios

### v2.0 (Futuro)
- Cache con Redis
- Versionado de configuraciones
- Eventos de dominio (pub/sub)
- UI de administración en backoffice

### v3.0 (Largo plazo)
- Feature flags dinámicos
- A/B testing por tenant
- Configuraciones jerárquicas (tenant → organización → global)

## 📚 Referencias

- [Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture - Alistair Cockburn](https://alistair.cockburn.us/hexagonal-architecture/)
- [Domain-Driven Design - Eric Evans](https://www.domainlanguage.com/ddd/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

## 🤝 Contribución

Al contribuir a este servicio, asegúrate de:

1. **Respetar las capas**: No crear dependencias inversas
2. **Testear el dominio**: Lógica de negocio debe tener tests
3. **Mantener interfaces pequeñas**: Principio de segregación
4. **Documentar decisiones**: Actualizar este documento si cambias arquitectura
5. **Seguir convenciones Go**: gofmt, golint, etc.

---

**Última actualización**: 2026-02-03  
**Versión**: 1.0.0  
**Autor**: SaaS MT Team - Tienda Vecina
