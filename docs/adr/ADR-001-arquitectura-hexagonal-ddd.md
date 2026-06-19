---
adr: ADR-001
status: accepted
skills:
  implement:
    - dev/hexagonal-go
  verify:
    - dev/go-hex-audit
---
# ADR-001: Arquitectura Hexagonal y DDD

**Estado**: Aceptado
**Fecha**: 2026-02-03
**Contexto**: El Tenant Service es el centro de configuración operativa y fiscal de un SaaS multi-tenant. Las reglas de negocio (optimistic locking, validación de fiscal_mode, políticas de stock) deben estar aisladas de la infraestructura concreta (PostgreSQL, Gin, EventBus) para ser testeables y evolutivas.

## Decisión

Adoptamos arquitectura hexagonal (Ports & Adapters) con los principios de DDD. La estructura de capas es: `domain` (lógica pura sin dependencias externas) ← `application` (use cases con repositorios como interfaces) ← `infrastructure` (adaptadores concretos: HTTP controllers, PostgreSQL repositories, EventBus publisher).

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| MVC plano (handlers → SQL directo) | No testeable sin base de datos real; las reglas de negocio se dispersan entre handlers y queries SQL |
| CQRS con Event Sourcing completo | Overhead operativo excesivo para el scope del servicio; los eventos de dominio actuales no justifican un event store propio |

## Consecuencias

**Positivas**: Lógica de negocio testeable con mocks (sin DB); clara separación de ports/adapters; interfaces de repositorio en el dominio permiten cambiar implementaciones sin tocar use cases.

**Negativas / trade-offs**: Mayor cantidad de archivos y capas; curva de aprendizaje inicial para nuevos colaboradores.

**Neutral**: Los DTOs de respuesta (`application/response`) importan entidades de dominio para conversión — patrón assembler aceptado en la capa de aplicación.
