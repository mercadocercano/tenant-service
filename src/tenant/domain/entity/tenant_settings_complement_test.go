package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantSettings_Validate_InvalidInvoiceGeneration(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()
	settings.InvoiceGeneration = "INVALID"

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoice_generation must be")
}

func TestTenantSettings_Validate_AllInvoiceGenerations(t *testing.T) {
	mother := NewTenantSettingsMother()
	validOptions := []string{
		InvoiceGenerationManual,
		InvoiceGenerationAutoOnSale,
		InvoiceGenerationAutoOnConfirm,
	}

	for _, opt := range validOptions {
		t.Run(opt, func(t *testing.T) {
			settings := mother.Default()
			settings.InvoiceGeneration = opt
			err := settings.Validate()
			assert.NoError(t, err, "InvoiceGeneration %s should be valid", opt)
		})
	}
}

func TestTenantSettings_Validate_InvalidTaxRegime(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()
	settings.TaxRegime = "INVALID"

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tax_regime must be")
}

func TestTenantSettings_Validate_AllTaxRegimes(t *testing.T) {
	mother := NewTenantSettingsMother()
	validRegimes := []string{TaxRegimeMonotributo, TaxRegimeResponsableInscripto}

	for _, regime := range validRegimes {
		t.Run(regime, func(t *testing.T) {
			settings := mother.Default()
			settings.TaxRegime = regime
			err := settings.Validate()
			assert.NoError(t, err, "TaxRegime %s should be valid", regime)
		})
	}
}

func TestTenantSettings_Validate_InvalidExchangeRateSource(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()
	settings.ExchangeRateSource = "INVALID"

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exchange_rate_source must be")
}

func TestTenantSettings_Validate_AllExchangeRateSources(t *testing.T) {
	mother := NewTenantSettingsMother()
	validSources := []string{ExchangeRateSourceManual, ExchangeRateSourceExternalAPI}

	for _, source := range validSources {
		t.Run(source, func(t *testing.T) {
			settings := mother.Default()
			settings.ExchangeRateSource = source
			err := settings.Validate()
			assert.NoError(t, err, "ExchangeRateSource %s should be valid", source)
		})
	}
}

func TestTenantSettings_Validate_NegativeDefaultCreditDays(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()
	settings.DefaultCreditDays = -1

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_credit_days must be >= 0")
}

func TestTenantSettings_Validate_NilTenantID(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Invalid_NilTenant()

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id cannot be nil")
}

func TestTenantSettings_Validate_NilCashCustomerID(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Invalid_NilCashCustomer()

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cash_customer_id cannot be nil")
}

func TestTenantSettings_Validate_FiscalRequired_Valid(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.WithFiscalRequired()

	err := settings.Validate()
	assert.NoError(t, err)
}

func TestTenantSettings_Validate_CreditEnabled_Valid(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.WithCreditEnabled()

	err := settings.Validate()
	assert.NoError(t, err)
	assert.True(t, settings.CreditEnabled)
	assert.Equal(t, 30, settings.DefaultCreditDays)
	assert.Equal(t, 50000.0, settings.MaxCreditLimit)
}

func TestTenantSettings_Validate_MultipleCurrencies_Valid(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.WithMultipleCurrencies()

	err := settings.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "USD", settings.BaseCurrency)
	assert.Contains(t, settings.AllowedCurrencies, "USD")
}

func TestTenantSettings_Update_ChangesAllFields(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()
	newCashCustomerID := uuid.New()

	settings.Update(
		"USD",
		[]string{"USD", "ARS"},
		ExchangeRateSourceExternalAPI,
		true,
		FiscalModeRequired,
		InvoiceGenerationAutoOnConfirm,
		false,
		true,
		true,
		"A",
		TaxRegimeResponsableInscripto,
		StockPolicyDeduct,
		false,
		true,
		true,
		45,
		75000.0,
		true,
		newCashCustomerID,
	)

	assert.Equal(t, "USD", settings.BaseCurrency)
	assert.Equal(t, []string{"USD", "ARS"}, settings.AllowedCurrencies)
	assert.Equal(t, ExchangeRateSourceExternalAPI, settings.ExchangeRateSource)
	assert.True(t, settings.AutoUpdateExchangeRate)
	assert.Equal(t, FiscalModeRequired, settings.FiscalMode)
	assert.Equal(t, InvoiceGenerationAutoOnConfirm, settings.InvoiceGeneration)
	assert.False(t, settings.AllowSaleIfAfipFails)
	assert.True(t, settings.AutoRetryFailedInvoices)
	assert.True(t, settings.EmailInvoiceAfterSuccess)
	assert.Equal(t, "A", settings.DefaultInvoiceType)
	assert.Equal(t, TaxRegimeResponsableInscripto, settings.TaxRegime)
	assert.Equal(t, StockPolicyDeduct, settings.StockPolicy)
	assert.False(t, settings.AllowNegativeStock)
	assert.True(t, settings.RequireStockValidationBeforeSale)
	assert.True(t, settings.CreditEnabled)
	assert.Equal(t, 45, settings.DefaultCreditDays)
	assert.Equal(t, 75000.0, settings.MaxCreditLimit)
	assert.True(t, settings.AllowSaleOverCreditLimit)
	assert.Equal(t, newCashCustomerID, settings.CashCustomerID)
	assert.False(t, settings.UpdatedAt.IsZero())
}

func TestTenantSettings_Update_DoesNotChangeVersion(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()
	require.Equal(t, 1, settings.Version)

	settings.Update(
		"ARS", []string{"ARS"}, ExchangeRateSourceManual, false,
		FiscalModeDisabled, InvoiceGenerationManual, true, false, false,
		"B", TaxRegimeMonotributo, StockPolicyIgnore, true, false,
		false, 30, 0, false, uuid.New(),
	)

	assert.Equal(t, 1, settings.Version, "Update no debe cambiar Version")
}

func TestTenantSettings_IncrementVersion_Multiple(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()

	for i := 0; i < 5; i++ {
		settings.IncrementVersion()
	}

	assert.Equal(t, 6, settings.Version)
}

func TestTenantSettings_NewDefaults(t *testing.T) {
	mother := NewTenantSettingsMother()
	settings := mother.Default()

	assert.Equal(t, ExchangeRateSourceManual, settings.ExchangeRateSource)
	assert.False(t, settings.AutoUpdateExchangeRate)
	assert.Equal(t, FiscalModeDisabled, settings.FiscalMode)
	assert.Equal(t, InvoiceGenerationManual, settings.InvoiceGeneration)
	assert.True(t, settings.AllowSaleIfAfipFails)
	assert.False(t, settings.AutoRetryFailedInvoices)
	assert.False(t, settings.EmailInvoiceAfterSuccess)
	assert.Equal(t, "B", settings.DefaultInvoiceType)
	assert.Equal(t, TaxRegimeMonotributo, settings.TaxRegime)
	assert.Equal(t, StockPolicyIgnore, settings.StockPolicy)
	assert.True(t, settings.AllowNegativeStock)
	assert.False(t, settings.RequireStockValidationBeforeSale)
	assert.False(t, settings.CreditEnabled)
	assert.Equal(t, 30, settings.DefaultCreditDays)
	assert.Equal(t, 0.0, settings.MaxCreditLimit)
	assert.False(t, settings.AllowSaleOverCreditLimit)
}
