-- Migration: Create tenant_settings table
-- Description: Configuraciones core estructuradas por tenant
-- Author: System
-- Date: 2026-02-19

-- ============================================
-- Create tenant_settings table
-- ============================================

CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id UUID PRIMARY KEY,

    -- MONETARIA
    base_currency VARCHAR(3) NOT NULL,
    allowed_currencies JSONB NOT NULL,
    exchange_rate_source VARCHAR(50) NOT NULL,
    auto_update_exchange_rate BOOLEAN NOT NULL,

    -- FISCAL
    fiscal_mode VARCHAR(50) NOT NULL,
    invoice_generation VARCHAR(50) NOT NULL,
    allow_sale_if_afip_fails BOOLEAN NOT NULL,
    auto_retry_failed_invoices BOOLEAN NOT NULL,
    email_invoice_after_success BOOLEAN NOT NULL,
    default_invoice_type VARCHAR(10) NOT NULL,
    tax_regime VARCHAR(50) NOT NULL,

    -- STOCK
    stock_policy VARCHAR(50) NOT NULL,
    allow_negative_stock BOOLEAN NOT NULL,
    require_stock_validation_before_sale BOOLEAN NOT NULL,

    -- CRÉDITO
    credit_enabled BOOLEAN NOT NULL,
    default_credit_days INT NOT NULL,
    max_credit_limit DECIMAL(18,2) NOT NULL,
    allow_sale_over_credit_limit BOOLEAN NOT NULL,

    -- CLIENTE CONTADO
    cash_customer_id UUID NOT NULL,

    -- CONTROL
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================
-- Create indexes
-- ============================================

CREATE INDEX IF NOT EXISTS idx_tenant_settings_tenant 
    ON tenant_settings(tenant_id);

-- ============================================
-- Comments
-- ============================================

COMMENT ON TABLE tenant_settings IS 'Configuraciones core estructuradas por tenant (modelo híbrido)';
COMMENT ON COLUMN tenant_settings.tenant_id IS 'Identificador único del tenant (PK)';
COMMENT ON COLUMN tenant_settings.base_currency IS 'Moneda base del tenant (ej: ARS, USD)';
COMMENT ON COLUMN tenant_settings.allowed_currencies IS 'Monedas permitidas en formato JSON array';
COMMENT ON COLUMN tenant_settings.exchange_rate_source IS 'Fuente de tasa de cambio: MANUAL | EXTERNAL_API';
COMMENT ON COLUMN tenant_settings.fiscal_mode IS 'Modo fiscal: DISABLED | OPTIONAL | REQUIRED';
COMMENT ON COLUMN tenant_settings.invoice_generation IS 'Generación de facturas: MANUAL | AUTO_ON_SALE | AUTO_ON_CONFIRM';
COMMENT ON COLUMN tenant_settings.stock_policy IS 'Política de stock: IGNORE | RESERVE | DEDUCT';
COMMENT ON COLUMN tenant_settings.tax_regime IS 'Régimen fiscal: MONOTRIBUTO | RESPONSABLE_INSCRIPTO';
COMMENT ON COLUMN tenant_settings.cash_customer_id IS 'ID del cliente genérico para ventas al contado';
COMMENT ON COLUMN tenant_settings.version IS 'Versión para optimistic locking';
COMMENT ON COLUMN tenant_settings.updated_at IS 'Timestamp de última actualización';
