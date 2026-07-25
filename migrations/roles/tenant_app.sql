-- migrations/roles/tenant_app.sql — PLAT-E29 T3 (RULE-09/RULE-10: rol de aplicación least-privilege, NOBYPASSRLS)
--
-- tenant-service corre hoy en runtime como `postgres` (superuser → BYPASSRLS: FORCE ROW LEVEL
-- SECURITY no aplica a superusers, así que la RLS de la migración 005 quedaría "activa" pero nunca
-- ejercida). Este rol least-privilege lo reemplaza en runtime en AMBAS conexiones (el flip del
-- compose se DIFIERE a T7 — flipear antes de T5/T6 rompe toda lectura, el repo aún no fija el GUC):
--
--   • SELECT/INSERT/UPDATE sobre tenant_config y tenant_settings — SIN DELETE: los métodos
--     Delete/GetAllByTenant (config) fueron podados en T4 (D2). config.Save = upsert
--     ON CONFLICT; settings.Save = update-then-insert (optimistic locking).
--   • SELECT/INSERT sobre points_of_sale — SIN UPDATE/DELETE: GetByID/Update fueron podados en
--     T4 (D2); solo quedan Create/ListByTenant/ListActiveByTenant.
--   • SELECT sobre schema_migrations (el runtime solo chequea la versión al arrancar, no aplica
--     DDL nuevo bajo este rol).
--   • eventbus (segunda conexión, DB compartida `eventbus`): D3 RESUELTO en la revisión L4 del
--     plan (must-fix A1 @dev-architect + @dev-security, 2026-07-21) — tenant-service SOLO
--     PUBLICA (`cmd/api/main.go` instancia únicamente `NewSQLEventStore`+`NewPublishEventUseCase`,
--     nunca `NewEventWorker`/`NewProcessEventUseCase`); `PublishEventUseCase.Execute` →
--     `SQLEventStore.Save` es un `INSERT INTO event_bus` puro, nunca toca `event_consumers`. NO
--     replica el precedente E24 (`ledger_app` consume además de publicar — caso estructuralmente
--     distinto, requiere DML completo). Grant final: INSERT ÚNICAMENTE sobre event_bus, CERO
--     grants sobre event_consumers — ni siquiera SELECT sobre event_bus (Save no lee; beneficio de
--     seguridad adicional: tenant_app no puede leer eventos ajenos en la tabla compartida, que no
--     tiene RLS propia).
--   • NOBYPASSRLS: la policy tenant_isolation de la migración 005 se ejerce de verdad.
--
-- NO es una migración numerada del boot: vive en migrations/roles/ (fuera del alcance de
-- `//go:embed migrations/*.sql` en migrations.go, glob NO recursivo → este archivo no entra al FS
-- embebido) para que NO se auto-aplique en el arranque bajo tenant_app (NOCREATEROLE no podría
-- crear roles). Se aplica UNA sola vez, out-of-band, contra el cluster por un admin (postgres).
-- Idempotente. Dos bloques — el primero conectado a tenant_db, el segundo a eventbus (`\connect`,
-- ver Contexto de PLAT-E29 T3): GRANT sobre tablas solo afecta la DB donde corre la sesión;
-- CREATE ROLE es cluster-wide y solo se ejecuta una vez (el `DO $$ IF NOT EXISTS` de abajo lo
-- hace idempotente si el archivo se corre más de una vez, incluso repetido contra cada DB).
--
-- Seguridad (mismo patrón que customer_app/E27, onboarding_app/E28): este script NO fija
-- password —un literal quedaría en el git history para siempre—. El rol se crea con LOGIN pero
-- SIN password (login efectivo deshabilitado hasta setearla); la password real se fija
-- out-of-band vía `ALTER ROLE tenant_app PASSWORD '...'` corrido a mano contra lab-postgres, nunca
-- versionada (dev local: tenant_app123, ver .env.example / docker-compose TENANT_DB_PASSWORD).
--
-- Aplicar: psql ... -f tenant_app.sql (NO con --single-transaction: el \connect intermedio abre
-- una sesión nueva, incompatible con una transacción única cross-DB; cada bloque es idempotente
-- por sí solo vía ON_ERROR_STOP=1).

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tenant_app') THEN
    CREATE ROLE tenant_app LOGIN
      NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
  END IF;
END
$$;

-- tenant_db
GRANT CONNECT ON DATABASE tenant_db TO tenant_app;
GRANT USAGE ON SCHEMA public TO tenant_app;

-- Hardening least-privilege (B1, serie E27): revocar privilegios de tabla que PUBLIC pudiera
-- conceder implícitamente; el rol solo obtiene lo que se le concede explícito abajo.
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;

-- Superficie VIVA post-poda T4 (D2): sin DELETE en ninguna tabla; sin UPDATE en points_of_sale.
GRANT SELECT, INSERT, UPDATE ON tenant_config TO tenant_app;    -- Save = upsert ON CONFLICT
GRANT SELECT, INSERT, UPDATE ON tenant_settings TO tenant_app;  -- Save = update-then-insert
GRANT SELECT, INSERT ON points_of_sale TO tenant_app;           -- Create + List (Update podado T4)

-- El runtime solo chequea la versión al arrancar (no aplica DDL nuevo → SELECT alcanza).
GRANT SELECT ON schema_migrations TO tenant_app;

-- eventbus (segunda conexión — D3 RESUELTO, must-fix A1): tenant-service SOLO PUBLICA, Save() es
-- INSERT puro, nunca toca event_consumers. CERO grants sobre event_consumers; ni SELECT sobre
-- event_bus (Save no lee).
\connect eventbus
GRANT CONNECT ON DATABASE eventbus TO tenant_app;
GRANT USAGE ON SCHEMA public TO tenant_app;
GRANT INSERT ON event_bus TO tenant_app;
-- sin grants sobre event_consumers
