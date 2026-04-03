package entity

import (
	"github.com/google/uuid"
)

// --- TenantConfigMother ---

// TenantConfigMother facilita la creación de TenantConfig para tests
type TenantConfigMother struct{}

func NewTenantConfigMother() *TenantConfigMother {
	return &TenantConfigMother{}
}

func (m *TenantConfigMother) Default() *TenantConfig {
	return NewTenantConfig(uuid.New(), "catalog.stock_policy", "REQUIRE_STOCK")
}

func (m *TenantConfigMother) WithKey(key, value string) *TenantConfig {
	return NewTenantConfig(uuid.New(), key, value)
}

func (m *TenantConfigMother) WithTenant(tenantID uuid.UUID) *TenantConfig {
	return NewTenantConfig(tenantID, "catalog.stock_policy", "REQUIRE_STOCK")
}

func (m *TenantConfigMother) WithTenantAndKey(tenantID uuid.UUID, key, value string) *TenantConfig {
	return NewTenantConfig(tenantID, key, value)
}

// --- PointOfSaleMother ---

// PointOfSaleMother facilita la creación de PointOfSale para tests
type PointOfSaleMother struct{}

func NewPointOfSaleMother() *PointOfSaleMother {
	return &PointOfSaleMother{}
}

func (m *PointOfSaleMother) Default() *PointOfSale {
	return NewPointOfSale(uuid.New(), 1, "Sucursal Central", true, "B")
}

func (m *PointOfSaleMother) WithTenant(tenantID uuid.UUID) *PointOfSale {
	return NewPointOfSale(tenantID, 1, "Sucursal Central", true, "B")
}

func (m *PointOfSaleMother) WithCode(code int) *PointOfSale {
	return NewPointOfSale(uuid.New(), code, "Sucursal Central", true, "B")
}

func (m *PointOfSaleMother) WithFiscalDisabled() *PointOfSale {
	return NewPointOfSale(uuid.New(), 1, "Sucursal Sin Fiscal", false, "C")
}

func (m *PointOfSaleMother) Invalid_NilTenant() *PointOfSale {
	return NewPointOfSale(uuid.Nil, 1, "Sucursal", true, "B")
}

func (m *PointOfSaleMother) Invalid_ZeroCode() *PointOfSale {
	return NewPointOfSale(uuid.New(), 0, "Sucursal", true, "B")
}

func (m *PointOfSaleMother) Invalid_EmptyDescription() *PointOfSale {
	return NewPointOfSale(uuid.New(), 1, "", true, "B")
}

func (m *PointOfSaleMother) Invalid_EmptyInvoiceType() *PointOfSale {
	return NewPointOfSale(uuid.New(), 1, "Sucursal", true, "")
}

// --- TenantSettingsMother ---

// TenantSettingsMother facilita la creación de TenantSettings para tests
type TenantSettingsMother struct{}

func NewTenantSettingsMother() *TenantSettingsMother {
	return &TenantSettingsMother{}
}

func (m *TenantSettingsMother) Default() *TenantSettings {
	return NewTenantSettings(uuid.New(), uuid.New())
}

func (m *TenantSettingsMother) WithTenant(tenantID uuid.UUID) *TenantSettings {
	return NewTenantSettings(tenantID, uuid.New())
}

func (m *TenantSettingsMother) WithCashCustomer(tenantID, cashCustomerID uuid.UUID) *TenantSettings {
	return NewTenantSettings(tenantID, cashCustomerID)
}

func (m *TenantSettingsMother) WithFiscalRequired() *TenantSettings {
	s := NewTenantSettings(uuid.New(), uuid.New())
	s.FiscalMode = FiscalModeRequired
	s.InvoiceGeneration = InvoiceGenerationAutoOnSale
	s.TaxRegime = TaxRegimeResponsableInscripto
	return s
}

func (m *TenantSettingsMother) WithCreditEnabled() *TenantSettings {
	s := NewTenantSettings(uuid.New(), uuid.New())
	s.CreditEnabled = true
	s.DefaultCreditDays = 30
	s.MaxCreditLimit = 50000
	return s
}

func (m *TenantSettingsMother) WithMultipleCurrencies() *TenantSettings {
	s := NewTenantSettings(uuid.New(), uuid.New())
	s.BaseCurrency = "USD"
	s.AllowedCurrencies = []string{"USD", "ARS", "EUR"}
	s.ExchangeRateSource = ExchangeRateSourceExternalAPI
	s.AutoUpdateExchangeRate = true
	return s
}

func (m *TenantSettingsMother) Invalid_NilTenant() *TenantSettings {
	return NewTenantSettings(uuid.Nil, uuid.New())
}

func (m *TenantSettingsMother) Invalid_NilCashCustomer() *TenantSettings {
	return NewTenantSettings(uuid.New(), uuid.Nil)
}
