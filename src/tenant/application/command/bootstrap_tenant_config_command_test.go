package command

import (
	"context"
	"errors"
	"testing"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestBootstrapTenantConfig_NewTenant prueba el bootstrap para un tenant nuevo
func TestBootstrapTenantConfig_NewTenant(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantConfigRepository)

	// Simular que no existen configuraciones previas
	mockRepo.On("GetByKey", ctx, tenantID, "catalog.stock_policy").Return(nil, false, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantConfig")).Return(nil)

	command := NewBootstrapTenantConfigCommand(mockRepo)

	// Act
	createdCount, err := command.Execute(ctx, tenantID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, createdCount, "Should create 1 config for new tenant")
	mockRepo.AssertExpectations(t)
}

// TestBootstrapTenantConfig_ExistingTenant prueba la idempotencia
func TestBootstrapTenantConfig_ExistingTenant(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantConfigRepository)

	// Simular que ya existe la configuración
	existingConfig := entity.NewTenantConfig(tenantID, "catalog.stock_policy", "REQUIRE_STOCK")
	mockRepo.On("GetByKey", ctx, tenantID, "catalog.stock_policy").Return(existingConfig, true, nil)
	// No debe llamar a Save si ya existe

	command := NewBootstrapTenantConfigCommand(mockRepo)

	// Act
	createdCount, err := command.Execute(ctx, tenantID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, createdCount, "Should not create any config for existing tenant")
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// TestBootstrapTenantConfig_DoubleCall prueba llamar bootstrap dos veces
func TestBootstrapTenantConfig_DoubleCall(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantConfigRepository)

	// Primera llamada: no existe
	mockRepo.On("GetByKey", ctx, tenantID, "catalog.stock_policy").Return(nil, false, nil).Once()
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantConfig")).Return(nil).Once()

	// Segunda llamada: ya existe
	existingConfig := entity.NewTenantConfig(tenantID, "catalog.stock_policy", "REQUIRE_STOCK")
	mockRepo.On("GetByKey", ctx, tenantID, "catalog.stock_policy").Return(existingConfig, true, nil).Once()

	command := NewBootstrapTenantConfigCommand(mockRepo)

	// Act - Primera llamada
	createdCount1, err1 := command.Execute(ctx, tenantID)

	// Assert - Primera llamada
	assert.NoError(t, err1)
	assert.Equal(t, 1, createdCount1, "First call should create 1 config")

	// Act - Segunda llamada (idempotencia)
	createdCount2, err2 := command.Execute(ctx, tenantID)

	// Assert - Segunda llamada
	assert.NoError(t, err2)
	assert.Equal(t, 0, createdCount2, "Second call should not create any config (idempotent)")

	mockRepo.AssertExpectations(t)
}

// TestBootstrapTenantConfig_RepositoryError prueba manejo de errores
func TestBootstrapTenantConfig_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantConfigRepository)

	// Simular error en GetByKey
	mockRepo.On("GetByKey", ctx, tenantID, "catalog.stock_policy").Return(nil, false, errors.New("database error"))

	command := NewBootstrapTenantConfigCommand(mockRepo)

	// Act
	createdCount, err := command.Execute(ctx, tenantID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, createdCount)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

// TestBootstrapTenantConfig_SaveError prueba error al guardar
func TestBootstrapTenantConfig_SaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantConfigRepository)

	// Simular que no existe pero falla al guardar
	mockRepo.On("GetByKey", ctx, tenantID, "catalog.stock_policy").Return(nil, false, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entity.TenantConfig")).Return(errors.New("save error"))

	command := NewBootstrapTenantConfigCommand(mockRepo)

	// Act
	createdCount, err := command.Execute(ctx, tenantID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, createdCount)
	assert.Contains(t, err.Error(), "save error")
	mockRepo.AssertExpectations(t)
}
