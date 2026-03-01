package request

import "errors"

// CreatePointOfSaleRequest representa la petición para crear un punto de venta
type CreatePointOfSaleRequest struct {
	Code                int    `json:"code" binding:"required"`
	Description         string `json:"description" binding:"required"`
	IsFiscalEnabled     bool   `json:"is_fiscal_enabled"`
	DefaultInvoiceType  string `json:"default_invoice_type" binding:"required"`
}

// Validate valida el request
func (r *CreatePointOfSaleRequest) Validate() error {
	if r.Code <= 0 {
		return errors.New("code must be greater than 0")
	}

	if r.Description == "" {
		return errors.New("description is required")
	}

	if r.DefaultInvoiceType == "" {
		return errors.New("default_invoice_type is required")
	}

	return nil
}
