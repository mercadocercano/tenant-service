package command

import (
	"context"
	"errors"
	"tenant/src/tenant/domain/entity"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func validUpdateParams(tenantID, cashCustomerID uuid.UUID, version int) UpdateTenantSettingsParams {
	return UpdateTenantSettingsParams{
		TenantID:                         tenantID,
		Version:                          version,
		BaseCurrency:                     "ARS",
		AllowedCurrencies:                []string{"ARS", "USD"},
		ExchangeRateSource:               entity.ExchangeRateSourceManual,
		AutoUpdateExchangeRate:           false,
		FiscalMode:                       entity.FiscalModeOptional,
		InvoiceGeneration:                entity.InvoiceGenerationManual,
		AllowSaleIfAfipFails:             true,
		AutoRetryFailedInvoices:          false,
		EmailInvoiceAfterSuccess:         false,
		DefaultInvoiceType:               "B",
		TaxRegime:                        entity.TaxRegimeMonotributo,
		StockPolicy:                      entity.StockPolicyReserve,
		AllowNegativeStock:               false,
		RequireStockValidationBeforeSale: true,
		CreditEnabled:                    false,
		DefaultCreditDays:                30,
		MaxCreditLimit:                   0,
		AllowSaleOverCreditLimit:         false,
		CashCustomerID:                   cashCustomerID,
	}
}

func TestUpdateTenantSettings_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	mockRepo := new(MockTenantSettingsRepository)
	mockPublisher := new(MockEventPublisher)

	currentSettings := entity.NewTenantSettings(tenantID, cashCustomerID)
	params := validUpdateParams(tenantID, cashCustomerID, 1)

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(currentSettings, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantSettings")).Return(nil)
	mockPublisher.On("Publish", ctx, tenantID.String(), "tenant_settings", "tenant.settings.updated", mock.Anything, "tenant-service").Return(nil)

	cmd := NewUpdateTenantSettingsCommand(mockRepo, mockPublisher)

	// Act
	result, err := cmd.Execute(ctx, params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entity.StockPolicyReserve, result.StockPolicy)
	assert.Equal(t, entity.FiscalModeOptional, result.FiscalMode)
	assert.Equal(t, 2, result.Version)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestUpdateTenantSettings_Execute_VersionConflict(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	mockRepo := new(MockTenantSettingsRepository)
	mockPublisher := new(MockEventPublisher)

	currentSettings := entity.NewTenantSettings(tenantID, cashCustomerID)
	currentSettings.IncrementVersion() // version = 2

	params := validUpdateParams(tenantID, cashCustomerID, 1) // envía version 1

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(currentSettings, nil)

	cmd := NewUpdateTenantSettingsCommand(mockRepo, mockPublisher)

	// Act
	result, err := cmd.Execute(ctx, params)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "version conflict")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestUpdateTenantSettings_Execute_GetByTenantIDError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	mockRepo := new(MockTenantSettingsRepository)
	mockPublisher := new(MockEventPublisher)

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(nil, errors.New("db error"))

	params := validUpdateParams(tenantID, cashCustomerID, 1)
	cmd := NewUpdateTenantSettingsCommand(mockRepo, mockPublisher)

	// Act
	result, err := cmd.Execute(ctx, params)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
}

func TestUpdateTenantSettings_Execute_ValidationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	mockRepo := new(MockTenantSettingsRepository)
	mockPublisher := new(MockEventPublisher)

	currentSettings := entity.NewTenantSettings(tenantID, cashCustomerID)

	params := validUpdateParams(tenantID, cashCustomerID, 1)
	params.FiscalMode = "INVALID_MODE"

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(currentSettings, nil)

	cmd := NewUpdateTenantSettingsCommand(mockRepo, mockPublisher)

	// Act
	result, err := cmd.Execute(ctx, params)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "fiscal_mode must be")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestUpdateTenantSettings_Execute_SaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	mockRepo := new(MockTenantSettingsRepository)
	mockPublisher := new(MockEventPublisher)

	currentSettings := entity.NewTenantSettings(tenantID, cashCustomerID)
	params := validUpdateParams(tenantID, cashCustomerID, 1)

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(currentSettings, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantSettings")).Return(errors.New("save failed"))

	cmd := NewUpdateTenantSettingsCommand(mockRepo, mockPublisher)

	// Act
	result, err := cmd.Execute(ctx, params)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "save failed")
}

func TestUpdateTenantSettings_Execute_EventPublishError_DoesNotFail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()

	mockRepo := new(MockTenantSettingsRepository)
	mockPublisher := new(MockEventPublisher)

	currentSettings := entity.NewTenantSettings(tenantID, cashCustomerID)
	params := validUpdateParams(tenantID, cashCustomerID, 1)

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(currentSettings, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantSettings")).Return(nil)
	mockPublisher.On("Publish", ctx, tenantID.String(), "tenant_settings", "tenant.settings.updated", mock.Anything, "tenant-service").Return(errors.New("event bus down"))

	cmd := NewUpdateTenantSettingsCommand(mockRepo, mockPublisher)

	// Act
	result, err := cmd.Execute(ctx, params)

	// Assert: no debe fallar aunque el evento no se publique
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Version)
}
