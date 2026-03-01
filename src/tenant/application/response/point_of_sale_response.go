package response

import (
	"tenant/src/tenant/domain/entity"
	"time"
)

// PointOfSaleResponse representa la respuesta de un punto de venta
type PointOfSaleResponse struct {
	ID                 string `json:"id"`
	TenantID           string `json:"tenant_id"`
	Code               int    `json:"code"`
	Description        string `json:"description"`
	IsFiscalEnabled    bool   `json:"is_fiscal_enabled"`
	DefaultInvoiceType string `json:"default_invoice_type"`
	IsActive           bool   `json:"is_active"`
	CreatedAt          string `json:"created_at"`
	Version            int    `json:"version"`
}

// FromPointOfSale convierte la entidad a DTO de respuesta
func FromPointOfSale(pos *entity.PointOfSale) *PointOfSaleResponse {
	return &PointOfSaleResponse{
		ID:                 pos.ID.String(),
		TenantID:           pos.TenantID.String(),
		Code:               pos.Code,
		Description:        pos.Description,
		IsFiscalEnabled:    pos.IsFiscalEnabled,
		DefaultInvoiceType: pos.DefaultInvoiceType,
		IsActive:           pos.IsActive,
		CreatedAt:          pos.CreatedAt.Format(time.RFC3339),
		Version:            pos.Version,
	}
}

// FromPointOfSaleList convierte una lista de entidades a DTOs
func FromPointOfSaleList(posList []*entity.PointOfSale) []*PointOfSaleResponse {
	responses := make([]*PointOfSaleResponse, 0, len(posList))
	for _, pos := range posList {
		responses = append(responses, FromPointOfSale(pos))
	}
	return responses
}
