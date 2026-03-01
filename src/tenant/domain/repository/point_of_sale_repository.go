package repository

import (
	"context"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
)

// PointOfSaleRepository define el contrato para persistir puntos de venta
type PointOfSaleRepository interface {
	// Create crea un nuevo punto de venta
	Create(ctx context.Context, pos *entity.PointOfSale) error

	// GetByID obtiene un punto de venta por ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PointOfSale, error)

	// ListByTenant obtiene todos los puntos de venta de un tenant
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error)

	// ListActiveByTenant obtiene solo los puntos activos de un tenant
	ListActiveByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error)

	// Update actualiza un punto de venta existente
	Update(ctx context.Context, pos *entity.PointOfSale) error
}
