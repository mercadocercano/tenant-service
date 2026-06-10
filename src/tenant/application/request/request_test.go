package request

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// --- SetTenantConfigRequest ---

func TestSetTenantConfigRequest_Valid(t *testing.T) {
	r := &SetTenantConfigRequest{Key: "catalog.stock_policy", Value: "REQUIRE_STOCK"}
	assert.NoError(t, r.Validate())
}

func TestSetTenantConfigRequest_EmptyKey(t *testing.T) {
	r := &SetTenantConfigRequest{Key: "", Value: "REQUIRE_STOCK"}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key")
}

func TestSetTenantConfigRequest_EmptyValue(t *testing.T) {
	r := &SetTenantConfigRequest{Key: "catalog.stock_policy", Value: ""}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value")
}

func TestValidationError_Error(t *testing.T) {
	e := &ValidationError{Field: "key", Message: "key cannot be empty"}
	assert.Equal(t, "key cannot be empty", e.Error())
}

// --- CreatePointOfSaleRequest ---

func TestCreatePointOfSaleRequest_Valid(t *testing.T) {
	r := &CreatePointOfSaleRequest{Code: 1, Description: "Sucursal Central", DefaultInvoiceType: "B"}
	assert.NoError(t, r.Validate())
}

func TestCreatePointOfSaleRequest_ZeroCode(t *testing.T) {
	r := &CreatePointOfSaleRequest{Code: 0, Description: "Sucursal", DefaultInvoiceType: "B"}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code")
}

func TestCreatePointOfSaleRequest_NegativeCode(t *testing.T) {
	r := &CreatePointOfSaleRequest{Code: -1, Description: "Sucursal", DefaultInvoiceType: "B"}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code")
}

func TestCreatePointOfSaleRequest_EmptyDescription(t *testing.T) {
	r := &CreatePointOfSaleRequest{Code: 1, Description: "", DefaultInvoiceType: "B"}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

func TestCreatePointOfSaleRequest_EmptyInvoiceType(t *testing.T) {
	r := &CreatePointOfSaleRequest{Code: 1, Description: "Sucursal", DefaultInvoiceType: ""}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_invoice_type")
}

// --- UpdateTenantSettingsRequest ---

func validUpdateRequest() *UpdateTenantSettingsRequest {
	return &UpdateTenantSettingsRequest{
		Version:            1,
		BaseCurrency:       "ARS",
		AllowedCurrencies:  []string{"ARS"},
		ExchangeRateSource: "MANUAL",
		FiscalMode:         "DISABLED",
		InvoiceGeneration:  "MANUAL",
		DefaultInvoiceType: "B",
		TaxRegime:          "MONOTRIBUTO",
		StockPolicy:        "IGNORE",
		DefaultCreditDays:  30,
		MaxCreditLimit:     0,
		CashCustomerID:     uuid.New().String(),
	}
}

func TestUpdateTenantSettingsRequest_Valid(t *testing.T) {
	assert.NoError(t, validUpdateRequest().Validate())
}

func TestUpdateTenantSettingsRequest_ZeroVersion(t *testing.T) {
	r := validUpdateRequest()
	r.Version = 0
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestUpdateTenantSettingsRequest_EmptyBaseCurrency(t *testing.T) {
	r := validUpdateRequest()
	r.BaseCurrency = ""
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base_currency")
}

func TestUpdateTenantSettingsRequest_EmptyAllowedCurrencies(t *testing.T) {
	r := validUpdateRequest()
	r.AllowedCurrencies = []string{}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed_currencies")
}

func TestUpdateTenantSettingsRequest_NegativeCreditDays(t *testing.T) {
	r := validUpdateRequest()
	r.DefaultCreditDays = -1
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_credit_days")
}

func TestUpdateTenantSettingsRequest_NegativeCreditLimit(t *testing.T) {
	r := validUpdateRequest()
	r.MaxCreditLimit = -100
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_credit_limit")
}

func TestUpdateTenantSettingsRequest_InvalidCashCustomerID(t *testing.T) {
	r := validUpdateRequest()
	r.CashCustomerID = "not-a-uuid"
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cash_customer_id")
}
