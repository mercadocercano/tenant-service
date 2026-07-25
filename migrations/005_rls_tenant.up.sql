-- Migration: 005_rls_tenant.sql
-- Description: RLS fail-closed en tenant_config, tenant_settings y points_of_sale (enable +
-- force + policy tenant_isolation con USING y WITH CHECK, sin caso global, sin break_glass).
-- PLAT-E29 T2 — RLS retrofit RULE-09/RULE-10. El GRANT del rol de aplicación vive en el DDL
-- de roles (migrations/roles/), NO acá: las migraciones numeradas quedan role-agnostic.

ALTER TABLE tenant_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_config FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON tenant_config
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_settings FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON tenant_settings
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

ALTER TABLE points_of_sale ENABLE ROW LEVEL SECURITY;
ALTER TABLE points_of_sale FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON points_of_sale
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);
