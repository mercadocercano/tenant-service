package response

import (
	"tenant/src/tenant/domain/entity"
	"time"
)

// TenantSettingsResponse representa la respuesta de configuraciones
type TenantSettingsResponse struct {
	TenantID string `json:"tenant_id"`

	// MONETARIA
	BaseCurrency           string   `json:"base_currency"`
	AllowedCurrencies      []string `json:"allowed_currencies"`
	ExchangeRateSource     string   `json:"exchange_rate_source"`
	AutoUpdateExchangeRate bool     `json:"auto_update_exchange_rate"`

	// FISCAL
	FiscalMode               string `json:"fiscal_mode"`
	InvoiceGeneration        string `json:"invoice_generation"`
	AllowSaleIfAfipFails     bool   `json:"allow_sale_if_afip_fails"`
	AutoRetryFailedInvoices  bool   `json:"auto_retry_failed_invoices"`
	EmailInvoiceAfterSuccess bool   `json:"email_invoice_after_success"`
	DefaultInvoiceType       string `json:"default_invoice_type"`
	TaxRegime                string `json:"tax_regime"`

	// STOCK
	StockPolicy                      string `json:"stock_policy"`
	AllowNegativeStock               bool   `json:"allow_negative_stock"`
	RequireStockValidationBeforeSale bool   `json:"require_stock_validation_before_sale"`

	// CRÉDITO
	CreditEnabled            bool    `json:"credit_enabled"`
	DefaultCreditDays        int     `json:"default_credit_days"`
	MaxCreditLimit           float64 `json:"max_credit_limit"`
	AllowSaleOverCreditLimit bool    `json:"allow_sale_over_credit_limit"`

	// CLIENTE CONTADO
	CashCustomerID string `json:"cash_customer_id"`

	// CONTROL
	Version   int    `json:"version"`
	UpdatedAt string `json:"updated_at"`
}

// FromTenantSettings convierte la entidad a DTO de respuesta
func FromTenantSettings(settings *entity.TenantSettings) *TenantSettingsResponse {
	return &TenantSettingsResponse{
		TenantID:                         settings.TenantID.String(),
		BaseCurrency:                     settings.BaseCurrency,
		AllowedCurrencies:                settings.AllowedCurrencies,
		ExchangeRateSource:               settings.ExchangeRateSource,
		AutoUpdateExchangeRate:           settings.AutoUpdateExchangeRate,
		FiscalMode:                       settings.FiscalMode,
		InvoiceGeneration:                settings.InvoiceGeneration,
		AllowSaleIfAfipFails:             settings.AllowSaleIfAfipFails,
		AutoRetryFailedInvoices:          settings.AutoRetryFailedInvoices,
		EmailInvoiceAfterSuccess:         settings.EmailInvoiceAfterSuccess,
		DefaultInvoiceType:               settings.DefaultInvoiceType,
		TaxRegime:                        settings.TaxRegime,
		StockPolicy:                      settings.StockPolicy,
		AllowNegativeStock:               settings.AllowNegativeStock,
		RequireStockValidationBeforeSale: settings.RequireStockValidationBeforeSale,
		CreditEnabled:                    settings.CreditEnabled,
		DefaultCreditDays:                settings.DefaultCreditDays,
		MaxCreditLimit:                   settings.MaxCreditLimit,
		AllowSaleOverCreditLimit:         settings.AllowSaleOverCreditLimit,
		CashCustomerID:                   settings.CashCustomerID.String(),
		Version:                          settings.Version,
		UpdatedAt:                        settings.UpdatedAt.Format(time.RFC3339),
	}
}
