# ADR-003: Domain Events vía EventBus (PostgreSQL-backed)

**Estado**: Aceptado
**Fecha**: 2026-02-19
**Contexto**: Al actualizar `tenant_settings` o crear un `PointOfSale`, otros servicios del ecosistema (sales-service, fiscal-service, etc.) necesitan reaccionar al cambio. Se necesita un mecanismo de notificación desacoplado.

## Decisión

Publicamos eventos de dominio (`tenant.settings.updated`, `tenant.point_of_sale.created`) al finalizar exitosamente cada command, usando el `EventPublisherAdapter` que adapta el puerto de dominio al `eventbus` de la infraestructura. El eventbus usa PostgreSQL como backing store.

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| REST callbacks síncronos | Acoplamiento temporal fuerte; si el servicio receptor está caído, el comando del tenant falla |
| Base de datos compartida (shared DB) | Viola los límites de bounded context; los servicios quedan acoplados al schema de tenant |
| Kafka / RabbitMQ | Overhead operativo no justificado en la etapa actual del ecosistema |

## Consecuencias

**Positivas**: Desacoplamiento entre servicios; el `EventPublisherAdapter` implementa un puerto de dominio (`EventPublisher` interface) sin contaminar la lógica de aplicación con detalles de infraestructura.

**Negativas / trade-offs**: La publicación de eventos y la escritura de datos no son atómicas (ver ADR-002). Si el servicio falla entre el commit a `tenant_db` y la publicación en `eventbus`, el evento se pierde.

**Neutral**: Los commands verifican el resultado de la publicación y loguean warnings si falla — comportamiento best-effort intencional para no comprometer la escritura principal.
