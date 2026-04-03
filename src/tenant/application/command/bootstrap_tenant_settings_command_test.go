package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBootstrapTenantSettings_Execute_NewTenant(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	mockRepo.On("Exists", ctx, tenantID).Return(false, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantSettings")).Return(nil)

	cmd := NewBootstrapTenantSettingsCommand(mockRepo)

	// Act
	created, err := cmd.Execute(ctx, tenantID, cashCustomerID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, created)
	mockRepo.AssertExpectations(t)
}

func TestBootstrapTenantSettings_Execute_ExistingTenant(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	mockRepo.On("Exists", ctx, tenantID).Return(true, nil)

	cmd := NewBootstrapTenantSettingsCommand(mockRepo)

	// Act
	created, err := cmd.Execute(ctx, tenantID, cashCustomerID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, created)
	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestBootstrapTenantSettings_Execute_ExistsError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	mockRepo.On("Exists", ctx, tenantID).Return(false, errors.New("db connection error"))

	cmd := NewBootstrapTenantSettingsCommand(mockRepo)

	// Act
	created, err := cmd.Execute(ctx, tenantID, cashCustomerID)

	// Assert
	assert.Error(t, err)
	assert.False(t, created)
	assert.Contains(t, err.Error(), "db connection error")
}

func TestBootstrapTenantSettings_Execute_SaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	mockRepo.On("Exists", ctx, tenantID).Return(false, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantSettings")).Return(errors.New("save error"))

	cmd := NewBootstrapTenantSettingsCommand(mockRepo)

	// Act
	created, err := cmd.Execute(ctx, tenantID, cashCustomerID)

	// Assert
	assert.Error(t, err)
	assert.False(t, created)
	assert.Contains(t, err.Error(), "save error")
}

func TestBootstrapTenantSettings_Execute_Idempotent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	cashCustomerID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	// Primera llamada: no existe
	mockRepo.On("Exists", ctx, tenantID).Return(false, nil).Once()
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantSettings")).Return(nil).Once()

	// Segunda llamada: ya existe
	mockRepo.On("Exists", ctx, tenantID).Return(true, nil).Once()

	cmd := NewBootstrapTenantSettingsCommand(mockRepo)

	// Act
	created1, err1 := cmd.Execute(ctx, tenantID, cashCustomerID)
	created2, err2 := cmd.Execute(ctx, tenantID, cashCustomerID)

	// Assert
	assert.NoError(t, err1)
	assert.True(t, created1)
	assert.NoError(t, err2)
	assert.False(t, created2)
	mockRepo.AssertExpectations(t)
}
