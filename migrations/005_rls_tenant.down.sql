-- Migration: 005_rls_tenant (down)
-- Description: Revert RLS + policies (orden inverso al up)

DROP POLICY IF EXISTS tenant_isolation ON points_of_sale;
ALTER TABLE points_of_sale NO FORCE ROW LEVEL SECURITY;
ALTER TABLE points_of_sale DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON tenant_settings;
ALTER TABLE tenant_settings NO FORCE ROW LEVEL SECURITY;
ALTER TABLE tenant_settings DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON tenant_config;
ALTER TABLE tenant_config NO FORCE ROW LEVEL SECURITY;
ALTER TABLE tenant_config DISABLE ROW LEVEL SECURITY;
