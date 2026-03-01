package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPointOfSale(t *testing.T) {
	tenantID := uuid.New()

	pos := NewPointOfSale(tenantID, 1, "Sucursal Centro", true, "B")

	assert.NotEqual(t, uuid.Nil, pos.ID)
	assert.Equal(t, tenantID, pos.TenantID)
	assert.Equal(t, 1, pos.Code)
	assert.Equal(t, "Sucursal Centro", pos.Description)
	assert.True(t, pos.IsFiscalEnabled)
	assert.Equal(t, "B", pos.DefaultInvoiceType)
	assert.True(t, pos.IsActive)
	assert.Equal(t, 1, pos.Version)
}

func TestPointOfSale_Validate_Success(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 1, "Sucursal Centro", true, "B")

	err := pos.Validate()
	assert.NoError(t, err)
}

func TestPointOfSale_Validate_InvalidCode(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 0, "Sucursal Centro", true, "B")

	err := pos.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code must be greater than 0")
}

func TestPointOfSale_Validate_EmptyDescription(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 1, "", true, "B")

	err := pos.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description cannot be empty")
}

func TestPointOfSale_Validate_EmptyInvoiceType(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 1, "Sucursal Centro", true, "")

	err := pos.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_invoice_type cannot be empty")
}

func TestPointOfSale_Activate(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 1, "Sucursal Centro", true, "B")
	pos.IsActive = false
	originalVersion := pos.Version

	pos.Activate()

	assert.True(t, pos.IsActive)
	assert.Equal(t, originalVersion+1, pos.Version)
}

func TestPointOfSale_Deactivate(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 1, "Sucursal Centro", true, "B")
	originalVersion := pos.Version

	pos.Deactivate()

	assert.False(t, pos.IsActive)
	assert.Equal(t, originalVersion+1, pos.Version)
}

func TestPointOfSale_Update(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 1, "Sucursal Centro", true, "B")
	originalVersion := pos.Version

	pos.Update("Sucursal Norte", false, "A")

	assert.Equal(t, "Sucursal Norte", pos.Description)
	assert.False(t, pos.IsFiscalEnabled)
	assert.Equal(t, "A", pos.DefaultInvoiceType)
	assert.Equal(t, originalVersion+1, pos.Version)
}
