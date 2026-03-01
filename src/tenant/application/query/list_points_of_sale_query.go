package query

import (
	"context"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// ListPointsOfSaleQuery representa el caso de uso para listar puntos de venta
type ListPointsOfSaleQuery struct {
	repository repository.PointOfSaleRepository
}

// NewListPointsOfSaleQuery crea una nueva instancia del query
func NewListPointsOfSaleQuery(repo repository.PointOfSaleRepository) *ListPointsOfSaleQuery {
	return &ListPointsOfSaleQuery{
		repository: repo,
	}
}

// Execute ejecuta el query
func (q *ListPointsOfSaleQuery) Execute(ctx context.Context, tenantID uuid.UUID, onlyActive bool) ([]*entity.PointOfSale, error) {
	if onlyActive {
		return q.repository.ListActiveByTenant(ctx, tenantID)
	}
	return q.repository.ListByTenant(ctx, tenantID)
}
