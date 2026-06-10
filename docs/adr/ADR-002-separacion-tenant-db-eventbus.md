# ADR-002: Separación de tenant_db y eventbus en PostgreSQL

**Estado**: Aceptado
**Fecha**: 2026-02-03
**Contexto**: El servicio necesita persistir configuración del tenant (`tenant_db`) y publicar eventos de dominio (`eventbus`). Se evaluó si usar una sola base de datos con tablas separadas o dos bases de datos distintas en la misma instancia PostgreSQL.

## Decisión

Usamos dos bases de datos PostgreSQL separadas en la misma instancia (`lab-postgres`): `tenant_db` para datos del dominio y `eventbus` para el bus de eventos. Cada una tiene su propio connection pool en el servicio.

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| Un solo schema con tablas separadas | Acoplamiento entre la escritura de datos del tenant y la escritura de eventos; una migración fallida afecta ambos |
| Tabla de outbox en tenant_db (Transactional Outbox) | Patrón más robusto pero requiere un componente adicional (poller/relay) para publicar eventos — overhead innecesario en la etapa actual |
| Kafka / RabbitMQ | Overhead operativo y de infraestructura; el lab usa eventbus basado en PostgreSQL que cubre los casos actuales |

## Consecuencias

**Positivas**: Aislamiento operativo — un fallo en la DB de eventos no afecta la escritura principal; los dos dominios (datos vs. eventos) evolucionan independientemente.

**Negativas / trade-offs**: No hay atomicidad entre la escritura del dato y la publicación del evento (no hay transaction que abarque dos DBs); en caso de fallo entre ambas operaciones puede quedar un estado inconsistente. Se acepta como trade-off conocido.

**Neutral**: El docker-compose crea ambas bases con un servicio `postgres-setup` one-shot antes de iniciar el servicio principal.
