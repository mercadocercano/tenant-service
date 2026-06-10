package response

import (
	"tenant/src/tenant/domain/entity"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newTenantConfig() *entity.TenantConfig {
	return entity.NewTenantConfig(uuid.New(), "catalog.stock_policy", "REQUIRE_STOCK")
}

func TestFromEntity_MapsAllFields(t *testing.T) {
	cfg := newTenantConfig()
	r := FromEntity(cfg)

	assert.Equal(t, cfg.ID, r.ID)
	assert.Equal(t, cfg.TenantID, r.TenantID)
	assert.Equal(t, cfg.Key, r.Key)
	assert.Equal(t, cfg.Value, r.Value)
	assert.Equal(t, cfg.CreatedAt, r.CreatedAt)
	assert.Equal(t, cfg.UpdatedAt, r.UpdatedAt)
}

func TestNewSimpleResponse_WithValue(t *testing.T) {
	v := "REQUIRE_STOCK"
	r := NewSimpleResponse("catalog.stock_policy", &v)

	assert.Equal(t, "catalog.stock_policy", r.Key)
	assert.NotNil(t, r.Value)
	assert.Equal(t, "REQUIRE_STOCK", *r.Value)
}

func TestNewSimpleResponse_NilValue(t *testing.T) {
	r := NewSimpleResponse("catalog.stock_policy", nil)

	assert.Equal(t, "catalog.stock_policy", r.Key)
	assert.Nil(t, r.Value)
}

func newPointOfSale() *entity.PointOfSale {
	return entity.NewPointOfSale(uuid.New(), 1, "Sucursal Central", true, "B")
}

func TestFromPointOfSale_MapsAllFields(t *testing.T) {
	pos := newPointOfSale()
	r := FromPointOfSale(pos)

	assert.Equal(t, pos.ID.String(), r.ID)
	assert.Equal(t, pos.TenantID.String(), r.TenantID)
	assert.Equal(t, pos.Code, r.Code)
	assert.Equal(t, pos.Description, r.Description)
	assert.Equal(t, pos.IsFiscalEnabled, r.IsFiscalEnabled)
	assert.Equal(t, pos.DefaultInvoiceType, r.DefaultInvoiceType)
	assert.Equal(t, pos.IsActive, r.IsActive)
	assert.Equal(t, pos.Version, r.Version)
	assert.Equal(t, pos.CreatedAt.Format(time.RFC3339), r.CreatedAt)
}

func TestFromPointOfSaleList_Empty(t *testing.T) {
	result := FromPointOfSaleList([]*entity.PointOfSale{})
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFromPointOfSaleList_MultipleItems(t *testing.T) {
	p1 := entity.NewPointOfSale(uuid.New(), 1, "Sucursal A", true, "B")
	p2 := entity.NewPointOfSale(uuid.New(), 2, "Sucursal B", false, "C")

	result := FromPointOfSaleList([]*entity.PointOfSale{p1, p2})

	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].Code)
	assert.Equal(t, 2, result[1].Code)
}

func newTenantSettings() *entity.TenantSettings {
	return entity.NewTenantSettings(uuid.New(), uuid.New())
}

func TestFromTenantSettings_MapsAllFields(t *testing.T) {
	s := newTenantSettings()
	r := FromTenantSettings(s)

	assert.Equal(t, s.TenantID.String(), r.TenantID)
	assert.Equal(t, s.BaseCurrency, r.BaseCurrency)
	assert.Equal(t, s.AllowedCurrencies, r.AllowedCurrencies)
	assert.Equal(t, s.ExchangeRateSource, r.ExchangeRateSource)
	assert.Equal(t, s.AutoUpdateExchangeRate, r.AutoUpdateExchangeRate)
	assert.Equal(t, s.FiscalMode, r.FiscalMode)
	assert.Equal(t, s.InvoiceGeneration, r.InvoiceGeneration)
	assert.Equal(t, s.AllowSaleIfAfipFails, r.AllowSaleIfAfipFails)
	assert.Equal(t, s.AutoRetryFailedInvoices, r.AutoRetryFailedInvoices)
	assert.Equal(t, s.EmailInvoiceAfterSuccess, r.EmailInvoiceAfterSuccess)
	assert.Equal(t, s.DefaultInvoiceType, r.DefaultInvoiceType)
	assert.Equal(t, s.TaxRegime, r.TaxRegime)
	assert.Equal(t, s.StockPolicy, r.StockPolicy)
	assert.Equal(t, s.AllowNegativeStock, r.AllowNegativeStock)
	assert.Equal(t, s.RequireStockValidationBeforeSale, r.RequireStockValidationBeforeSale)
	assert.Equal(t, s.CreditEnabled, r.CreditEnabled)
	assert.Equal(t, s.DefaultCreditDays, r.DefaultCreditDays)
	assert.Equal(t, s.MaxCreditLimit, r.MaxCreditLimit)
	assert.Equal(t, s.AllowSaleOverCreditLimit, r.AllowSaleOverCreditLimit)
	assert.Equal(t, s.CashCustomerID.String(), r.CashCustomerID)
	assert.Equal(t, s.Version, r.Version)
	assert.Equal(t, s.UpdatedAt.Format(time.RFC3339), r.UpdatedAt)
}
