-- Rollback de migrations/roles/tenant_app.sql — PLAT-E29 T3 (best-effort, dev local).
-- Revoca los grants en ambas DBs y elimina el rol. DROP ROLE falla si el rol es dueño de objetos
-- o tiene privilegios pendientes en otras DBs — best-effort, mismo patrón que customer_app/E27.
-- Aplicar: psql ... -f tenant_app.down.sql (mismo motivo que el up: NO --single-transaction, el
-- \connect abre una sesión nueva).

\connect eventbus
REVOKE ALL ON event_bus             FROM tenant_app;
REVOKE USAGE ON SCHEMA public       FROM tenant_app;
REVOKE CONNECT ON DATABASE eventbus FROM tenant_app;

\connect tenant_db
REVOKE ALL ON schema_migrations      FROM tenant_app;
REVOKE ALL ON points_of_sale         FROM tenant_app;
REVOKE ALL ON tenant_settings        FROM tenant_app;
REVOKE ALL ON tenant_config          FROM tenant_app;
REVOKE USAGE ON SCHEMA public        FROM tenant_app;
REVOKE CONNECT ON DATABASE tenant_db FROM tenant_app;
DROP ROLE IF EXISTS tenant_app;
