package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTenantSettings(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)

	assert.Equal(t, tenantID, settings.TenantID)
	assert.Equal(t, cashCustomerID, settings.CashCustomerID)
	assert.Equal(t, "ARS", settings.BaseCurrency)
	assert.Equal(t, []string{"ARS"}, settings.AllowedCurrencies)
	assert.Equal(t, FiscalModeDisabled, settings.FiscalMode)
	assert.Equal(t, StockPolicyIgnore, settings.StockPolicy)
	assert.Equal(t, 1, settings.Version)
	assert.False(t, settings.CreditEnabled)
}

func TestTenantSettings_Validate_Success(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)

	err := settings.Validate()
	assert.NoError(t, err)
}

func TestTenantSettings_Validate_InvalidFiscalMode(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)
	settings.FiscalMode = "INVALID"

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fiscal_mode must be")
}

func TestTenantSettings_Validate_InvalidStockPolicy(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)
	settings.StockPolicy = "INVALID"

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stock_policy must be")
}

func TestTenantSettings_Validate_BaseCurrencyNotInAllowed(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)
	settings.BaseCurrency = "USD"
	settings.AllowedCurrencies = []string{"ARS", "EUR"}

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base_currency must be in allowed_currencies")
}

func TestTenantSettings_Validate_NegativeCreditLimit(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)
	settings.MaxCreditLimit = -100

	err := settings.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_credit_limit must be >= 0")
}

func TestTenantSettings_Update(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)
	originalVersion := settings.Version

	// Update settings
	settings.Update(
		"USD",
		[]string{"USD", "ARS"},
		ExchangeRateSourceExternalAPI,
		true,
		FiscalModeRequired,
		InvoiceGenerationAutoOnSale,
		false,
		true,
		true,
		"A",
		TaxRegimeResponsableInscripto,
		StockPolicyReserve,
		false,
		true,
		true,
		60,
		100000.00,
		false,
		uuid.New(),
	)

	assert.Equal(t, "USD", settings.BaseCurrency)
	assert.Equal(t, FiscalModeRequired, settings.FiscalMode)
	assert.Equal(t, StockPolicyReserve, settings.StockPolicy)
	assert.True(t, settings.CreditEnabled)
	assert.Equal(t, 60, settings.DefaultCreditDays)
	assert.Equal(t, originalVersion, settings.Version) // Version no cambia hasta IncrementVersion
}

func TestTenantSettings_IncrementVersion(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	settings := NewTenantSettings(tenantID, cashCustomerID)
	assert.Equal(t, 1, settings.Version)

	settings.IncrementVersion()
	assert.Equal(t, 2, settings.Version)

	settings.IncrementVersion()
	assert.Equal(t, 3, settings.Version)
}

func TestTenantSettings_Validate_AllFiscalModes(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	validModes := []string{FiscalModeDisabled, FiscalModeOptional, FiscalModeRequired}

	for _, mode := range validModes {
		settings := NewTenantSettings(tenantID, cashCustomerID)
		settings.FiscalMode = mode
		err := settings.Validate()
		assert.NoError(t, err, "FiscalMode %s should be valid", mode)
	}
}

func TestTenantSettings_Validate_AllStockPolicies(t *testing.T) {
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	validPolicies := []string{StockPolicyIgnore, StockPolicyReserve, StockPolicyDeduct}

	for _, policy := range validPolicies {
		settings := NewTenantSettings(tenantID, cashCustomerID)
		settings.StockPolicy = policy
		err := settings.Validate()
		assert.NoError(t, err, "StockPolicy %s should be valid", policy)
	}
}
