package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPointOfSale_Validate_NilTenantID(t *testing.T) {
	mother := NewPointOfSaleMother()
	pos := mother.Invalid_NilTenant()

	err := pos.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id cannot be nil")
}

func TestPointOfSale_Validate_NegativeCode(t *testing.T) {
	pos := NewPointOfSale(uuid.New(), -1, "Sucursal", true, "B")

	err := pos.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code must be greater than 0")
}

func TestPointOfSale_NewPointOfSale_Defaults(t *testing.T) {
	mother := NewPointOfSaleMother()
	pos := mother.Default()

	assert.True(t, pos.IsActive, "Debe crearse activo por defecto")
	assert.Equal(t, 1, pos.Version, "Debe crearse con version 1")
	assert.False(t, pos.CreatedAt.IsZero(), "Debe tener timestamp de creacion")
}

func TestPointOfSale_Deactivate_ThenActivate(t *testing.T) {
	mother := NewPointOfSaleMother()
	pos := mother.Default()

	pos.Deactivate()
	assert.False(t, pos.IsActive)
	assert.Equal(t, 2, pos.Version)

	pos.Activate()
	assert.True(t, pos.IsActive)
	assert.Equal(t, 3, pos.Version)
}

func TestPointOfSale_Update_PreservesOtherFields(t *testing.T) {
	tenantID := uuid.New()
	pos := NewPointOfSale(tenantID, 5, "Sucursal Original", true, "B")
	originalID := pos.ID

	pos.Update("Sucursal Modificada", false, "A")

	assert.Equal(t, originalID, pos.ID, "ID no debe cambiar")
	assert.Equal(t, tenantID, pos.TenantID, "TenantID no debe cambiar")
	assert.Equal(t, 5, pos.Code, "Code no debe cambiar")
	assert.Equal(t, "Sucursal Modificada", pos.Description)
	assert.False(t, pos.IsFiscalEnabled)
	assert.Equal(t, "A", pos.DefaultInvoiceType)
}

func TestPointOfSale_WithFiscalDisabled(t *testing.T) {
	mother := NewPointOfSaleMother()
	pos := mother.WithFiscalDisabled()

	assert.False(t, pos.IsFiscalEnabled)
	assert.Equal(t, "C", pos.DefaultInvoiceType)
	assert.NoError(t, pos.Validate())
}

func TestPointOfSale_MultipleUpdates_IncrementsVersion(t *testing.T) {
	mother := NewPointOfSaleMother()
	pos := mother.Default()

	pos.Update("V2", true, "A")
	pos.Update("V3", false, "B")
	pos.Update("V4", true, "C")

	assert.Equal(t, 4, pos.Version)
	assert.Equal(t, "V4", pos.Description)
}
