-- Migration: Seed initial tenant configurations
-- Description: Datos iniciales de ejemplo para desarrollo
-- Author: System
-- Date: 2026-02-03

-- ============================================
-- Seed data (solo para desarrollo/testing)
-- ============================================

-- Ejemplo: Configuración de stock policy para un tenant de prueba
-- Nota: Este tenant debe existir en iam-service
-- INSERT INTO tenant_config (tenant_id, config_key, config_value)
-- VALUES 
--     ('00000000-0000-0000-0000-000000000001', 'catalog.stock_policy', 'IGNORE_STOCK'),
--     ('00000000-0000-0000-0000-000000000001', 'catalog.auto_publish', 'true');

-- Este archivo está comentado intencionalmente
-- Los datos reales se insertarán vía API o scripts de onboarding
