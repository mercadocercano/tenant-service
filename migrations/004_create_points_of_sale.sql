-- Migration: Create points_of_sale table
-- Description: Puntos de venta por tenant
-- Author: System
-- Date: 2026-02-19

-- ============================================
-- Create points_of_sale table
-- ============================================

CREATE TABLE IF NOT EXISTS points_of_sale (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code INT NOT NULL,
    description VARCHAR(255) NOT NULL,
    is_fiscal_enabled BOOLEAN NOT NULL DEFAULT true,
    default_invoice_type VARCHAR(10) NOT NULL DEFAULT 'B',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

-- ============================================
-- Create indexes
-- ============================================

CREATE INDEX IF NOT EXISTS idx_pos_tenant 
    ON points_of_sale(tenant_id);

CREATE INDEX IF NOT EXISTS idx_pos_tenant_active 
    ON points_of_sale(tenant_id, is_active);

CREATE INDEX IF NOT EXISTS idx_pos_code 
    ON points_of_sale(tenant_id, code);

-- ============================================
-- Comments
-- ============================================

COMMENT ON TABLE points_of_sale IS 'Puntos de venta por tenant';
COMMENT ON COLUMN points_of_sale.id IS 'Identificador único del punto de venta';
COMMENT ON COLUMN points_of_sale.tenant_id IS 'Identificador del tenant propietario';
COMMENT ON COLUMN points_of_sale.code IS 'Código numérico del punto de venta (ej: 1, 2, 3)';
COMMENT ON COLUMN points_of_sale.description IS 'Descripción del punto de venta';
COMMENT ON COLUMN points_of_sale.is_fiscal_enabled IS 'Si el punto está habilitado para facturación electrónica';
COMMENT ON COLUMN points_of_sale.default_invoice_type IS 'Tipo de factura por defecto (A, B, C, etc.)';
COMMENT ON COLUMN points_of_sale.is_active IS 'Si el punto de venta está activo';
COMMENT ON COLUMN points_of_sale.version IS 'Versión para optimistic locking';
