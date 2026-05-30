package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newPOSCmd(repo *MockPointOfSaleRepository) *CreatePointOfSaleCommand {
	publisher := new(MockEventPublisher)
	publisher.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return NewCreatePointOfSaleCommand(repo, publisher)
}

func TestCreatePointOfSale_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.PointOfSale")).Return(nil)

	cmd := newPOSCmd(mockRepo)

	// Act
	result, err := cmd.Execute(ctx, tenantID, 1, "Sucursal Centro", true, "B")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, 1, result.Code)
	assert.Equal(t, "Sucursal Centro", result.Description)
	assert.True(t, result.IsFiscalEnabled)
	assert.Equal(t, "B", result.DefaultInvoiceType)
	assert.True(t, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestCreatePointOfSale_Execute_ValidationError_ZeroCode(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	cmd := newPOSCmd(mockRepo)

	// Act
	result, err := cmd.Execute(ctx, tenantID, 0, "Sucursal", true, "B")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "code must be greater than 0")
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreatePointOfSale_Execute_ValidationError_EmptyDescription(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	cmd := newPOSCmd(mockRepo)

	// Act
	result, err := cmd.Execute(ctx, tenantID, 1, "", true, "B")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "description cannot be empty")
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreatePointOfSale_Execute_ValidationError_EmptyInvoiceType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	cmd := newPOSCmd(mockRepo)

	// Act
	result, err := cmd.Execute(ctx, tenantID, 1, "Sucursal", true, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "default_invoice_type cannot be empty")
}

func TestCreatePointOfSale_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.PointOfSale")).Return(errors.New("duplicate code"))

	cmd := newPOSCmd(mockRepo)

	// Act
	result, err := cmd.Execute(ctx, tenantID, 1, "Sucursal Centro", true, "B")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate code")
}

func TestCreatePointOfSale_Execute_WithFiscalDisabled(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.PointOfSale")).Return(nil)

	cmd := newPOSCmd(mockRepo)

	// Act
	result, err := cmd.Execute(ctx, tenantID, 2, "Sucursal Norte", false, "C")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsFiscalEnabled)
	assert.Equal(t, "C", result.DefaultInvoiceType)
}
