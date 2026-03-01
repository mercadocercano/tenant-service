package request

import (
	"errors"

	"github.com/google/uuid"
)

// UpdateTenantSettingsRequest representa la petición para actualizar configuraciones
type UpdateTenantSettingsRequest struct {
	Version int `json:"version" binding:"required"`

	// MONETARIA
	BaseCurrency           string   `json:"base_currency" binding:"required"`
	AllowedCurrencies      []string `json:"allowed_currencies" binding:"required"`
	ExchangeRateSource     string   `json:"exchange_rate_source" binding:"required"`
	AutoUpdateExchangeRate bool     `json:"auto_update_exchange_rate"`

	// FISCAL
	FiscalMode               string `json:"fiscal_mode" binding:"required"`
	InvoiceGeneration        string `json:"invoice_generation" binding:"required"`
	AllowSaleIfAfipFails     bool   `json:"allow_sale_if_afip_fails"`
	AutoRetryFailedInvoices  bool   `json:"auto_retry_failed_invoices"`
	EmailInvoiceAfterSuccess bool   `json:"email_invoice_after_success"`
	DefaultInvoiceType       string `json:"default_invoice_type" binding:"required"`
	TaxRegime                string `json:"tax_regime" binding:"required"`

	// STOCK
	StockPolicy                      string `json:"stock_policy" binding:"required"`
	AllowNegativeStock               bool   `json:"allow_negative_stock"`
	RequireStockValidationBeforeSale bool   `json:"require_stock_validation_before_sale"`

	// CRÉDITO
	CreditEnabled            bool    `json:"credit_enabled"`
	DefaultCreditDays        int     `json:"default_credit_days" binding:"required"`
	MaxCreditLimit           float64 `json:"max_credit_limit" binding:"required"`
	AllowSaleOverCreditLimit bool    `json:"allow_sale_over_credit_limit"`

	// CLIENTE CONTADO
	CashCustomerID string `json:"cash_customer_id" binding:"required"`
}

// Validate valida el request
func (r *UpdateTenantSettingsRequest) Validate() error {
	if r.Version <= 0 {
		return errors.New("version must be greater than 0")
	}

	if r.BaseCurrency == "" {
		return errors.New("base_currency is required")
	}

	if len(r.AllowedCurrencies) == 0 {
		return errors.New("allowed_currencies must have at least one currency")
	}

	if r.DefaultCreditDays < 0 {
		return errors.New("default_credit_days must be >= 0")
	}

	if r.MaxCreditLimit < 0 {
		return errors.New("max_credit_limit must be >= 0")
	}

	// Validar UUID de cash customer
	if _, err := uuid.Parse(r.CashCustomerID); err != nil {
		return errors.New("cash_customer_id must be a valid UUID")
	}

	return nil
}
