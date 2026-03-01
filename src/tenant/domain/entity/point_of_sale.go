package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PointOfSale representa un punto de venta del tenant
// Cada tenant puede tener múltiples puntos de venta (sucursales)
type PointOfSale struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	Code                int
	Description         string
	IsFiscalEnabled     bool
	DefaultInvoiceType  string
	IsActive            bool
	CreatedAt           time.Time
	Version             int
}

// NewPointOfSale crea una nueva instancia de punto de venta
func NewPointOfSale(
	tenantID uuid.UUID,
	code int,
	description string,
	isFiscalEnabled bool,
	defaultInvoiceType string,
) *PointOfSale {
	return &PointOfSale{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		Code:               code,
		Description:        description,
		IsFiscalEnabled:    isFiscalEnabled,
		DefaultInvoiceType: defaultInvoiceType,
		IsActive:           true,
		CreatedAt:          time.Now(),
		Version:            1,
	}
}

// Validate valida las reglas de negocio del punto de venta
func (pos *PointOfSale) Validate() error {
	if pos.TenantID == uuid.Nil {
		return errors.New("tenant_id cannot be nil")
	}

	if pos.Code <= 0 {
		return errors.New("code must be greater than 0")
	}

	if pos.Description == "" {
		return errors.New("description cannot be empty")
	}

	if pos.DefaultInvoiceType == "" {
		return errors.New("default_invoice_type cannot be empty")
	}

	return nil
}

// Activate activa el punto de venta
func (pos *PointOfSale) Activate() {
	pos.IsActive = true
	pos.Version++
}

// Deactivate desactiva el punto de venta
func (pos *PointOfSale) Deactivate() {
	pos.IsActive = false
	pos.Version++
}

// Update actualiza los campos del punto de venta
func (pos *PointOfSale) Update(description string, isFiscalEnabled bool, defaultInvoiceType string) {
	pos.Description = description
	pos.IsFiscalEnabled = isFiscalEnabled
	pos.DefaultInvoiceType = defaultInvoiceType
	pos.Version++
}
