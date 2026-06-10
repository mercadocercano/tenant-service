package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TenantSettings representa la configuración core estructurada de un tenant
// Es un agregado raíz que centraliza políticas operativas y fiscales
type TenantSettings struct {
	TenantID uuid.UUID

	// MONETARIA
	BaseCurrency           string
	AllowedCurrencies      []string
	ExchangeRateSource     string
	AutoUpdateExchangeRate bool

	// FISCAL
	FiscalMode               string
	InvoiceGeneration        string
	AllowSaleIfAfipFails     bool
	AutoRetryFailedInvoices  bool
	EmailInvoiceAfterSuccess bool
	DefaultInvoiceType       string
	TaxRegime                string

	// STOCK
	StockPolicy                      string
	AllowNegativeStock               bool
	RequireStockValidationBeforeSale bool

	// CRÉDITO
	CreditEnabled            bool
	DefaultCreditDays        int
	MaxCreditLimit           float64
	AllowSaleOverCreditLimit bool

	// CLIENTE CONTADO
	CashCustomerID uuid.UUID

	// CONTROL
	Version   int
	UpdatedAt time.Time
}

// Constantes de validación
const (
	// FiscalMode
	FiscalModeDisabled = "DISABLED"
	FiscalModeOptional = "OPTIONAL"
	FiscalModeRequired = "REQUIRED"

	// InvoiceGeneration
	InvoiceGenerationManual        = "MANUAL"
	InvoiceGenerationAutoOnSale    = "AUTO_ON_SALE"
	InvoiceGenerationAutoOnConfirm = "AUTO_ON_CONFIRM"

	// StockPolicy
	StockPolicyIgnore  = "IGNORE"
	StockPolicyReserve = "RESERVE"
	StockPolicyDeduct  = "DEDUCT"

	// TaxRegime
	TaxRegimeMonotributo          = "MONOTRIBUTO"
	TaxRegimeResponsableInscripto = "RESPONSABLE_INSCRIPTO"

	// ExchangeRateSource
	ExchangeRateSourceManual      = "MANUAL"
	ExchangeRateSourceExternalAPI = "EXTERNAL_API"
)

// NewTenantSettings crea una nueva instancia con valores por defecto seguros
func NewTenantSettings(tenantID uuid.UUID, cashCustomerID uuid.UUID) *TenantSettings {
	return &TenantSettings{
		TenantID:                         tenantID,
		BaseCurrency:                     "ARS",
		AllowedCurrencies:                []string{"ARS"},
		ExchangeRateSource:               ExchangeRateSourceManual,
		AutoUpdateExchangeRate:           false,
		FiscalMode:                       FiscalModeDisabled,
		InvoiceGeneration:                InvoiceGenerationManual,
		AllowSaleIfAfipFails:             true,
		AutoRetryFailedInvoices:          false,
		EmailInvoiceAfterSuccess:         false,
		DefaultInvoiceType:               "B",
		TaxRegime:                        TaxRegimeMonotributo,
		StockPolicy:                      StockPolicyIgnore,
		AllowNegativeStock:               true,
		RequireStockValidationBeforeSale: false,
		CreditEnabled:                    false,
		DefaultCreditDays:                30,
		MaxCreditLimit:                   0,
		AllowSaleOverCreditLimit:         false,
		CashCustomerID:                   cashCustomerID,
		Version:                          1,
		UpdatedAt:                        time.Now(),
	}
}

// Validate valida las reglas de negocio del agregado
func (ts *TenantSettings) Validate() error {
	// Validar fiscal_mode
	if !isValidFiscalMode(ts.FiscalMode) {
		return errors.New("fiscal_mode must be DISABLED, OPTIONAL, or REQUIRED")
	}

	// Validar invoice_generation
	if !isValidInvoiceGeneration(ts.InvoiceGeneration) {
		return errors.New("invoice_generation must be MANUAL, AUTO_ON_SALE, or AUTO_ON_CONFIRM")
	}

	// Validar stock_policy
	if !isValidStockPolicy(ts.StockPolicy) {
		return errors.New("stock_policy must be IGNORE, RESERVE, or DEDUCT")
	}

	// Validar tax_regime
	if !isValidTaxRegime(ts.TaxRegime) {
		return errors.New("tax_regime must be MONOTRIBUTO or RESPONSABLE_INSCRIPTO")
	}

	// Validar exchange_rate_source
	if !isValidExchangeRateSource(ts.ExchangeRateSource) {
		return errors.New("exchange_rate_source must be MANUAL or EXTERNAL_API")
	}

	// Validar base_currency está en allowed_currencies
	if !contains(ts.AllowedCurrencies, ts.BaseCurrency) {
		return errors.New("base_currency must be in allowed_currencies")
	}

	// Validar límites numéricos
	if ts.MaxCreditLimit < 0 {
		return errors.New("max_credit_limit must be >= 0")
	}

	if ts.DefaultCreditDays < 0 {
		return errors.New("default_credit_days must be >= 0")
	}

	// Validar UUIDs no vacíos
	if ts.TenantID == uuid.Nil {
		return errors.New("tenant_id cannot be nil")
	}

	if ts.CashCustomerID == uuid.Nil {
		return errors.New("cash_customer_id cannot be nil")
	}

	return nil
}

// Update actualiza los campos del settings manteniendo invariantes
func (ts *TenantSettings) Update(
	baseCurrency string,
	allowedCurrencies []string,
	exchangeRateSource string,
	autoUpdateExchangeRate bool,
	fiscalMode string,
	invoiceGeneration string,
	allowSaleIfAfipFails bool,
	autoRetryFailedInvoices bool,
	emailInvoiceAfterSuccess bool,
	defaultInvoiceType string,
	taxRegime string,
	stockPolicy string,
	allowNegativeStock bool,
	requireStockValidationBeforeSale bool,
	creditEnabled bool,
	defaultCreditDays int,
	maxCreditLimit float64,
	allowSaleOverCreditLimit bool,
	cashCustomerID uuid.UUID,
) {
	ts.BaseCurrency = baseCurrency
	ts.AllowedCurrencies = allowedCurrencies
	ts.ExchangeRateSource = exchangeRateSource
	ts.AutoUpdateExchangeRate = autoUpdateExchangeRate
	ts.FiscalMode = fiscalMode
	ts.InvoiceGeneration = invoiceGeneration
	ts.AllowSaleIfAfipFails = allowSaleIfAfipFails
	ts.AutoRetryFailedInvoices = autoRetryFailedInvoices
	ts.EmailInvoiceAfterSuccess = emailInvoiceAfterSuccess
	ts.DefaultInvoiceType = defaultInvoiceType
	ts.TaxRegime = taxRegime
	ts.StockPolicy = stockPolicy
	ts.AllowNegativeStock = allowNegativeStock
	ts.RequireStockValidationBeforeSale = requireStockValidationBeforeSale
	ts.CreditEnabled = creditEnabled
	ts.DefaultCreditDays = defaultCreditDays
	ts.MaxCreditLimit = maxCreditLimit
	ts.AllowSaleOverCreditLimit = allowSaleOverCreditLimit
	ts.CashCustomerID = cashCustomerID
	ts.UpdatedAt = time.Now()
}

// IncrementVersion incrementa la versión para optimistic locking
func (ts *TenantSettings) IncrementVersion() {
	ts.Version++
}

// Helper functions para validación

func isValidFiscalMode(mode string) bool {
	return mode == FiscalModeDisabled || mode == FiscalModeOptional || mode == FiscalModeRequired
}

func isValidInvoiceGeneration(gen string) bool {
	return gen == InvoiceGenerationManual || gen == InvoiceGenerationAutoOnSale || gen == InvoiceGenerationAutoOnConfirm
}

func isValidStockPolicy(policy string) bool {
	return policy == StockPolicyIgnore || policy == StockPolicyReserve || policy == StockPolicyDeduct
}

func isValidTaxRegime(regime string) bool {
	return regime == TaxRegimeMonotributo || regime == TaxRegimeResponsableInscripto
}

func isValidExchangeRateSource(source string) bool {
	return source == ExchangeRateSourceManual || source == ExchangeRateSourceExternalAPI
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
