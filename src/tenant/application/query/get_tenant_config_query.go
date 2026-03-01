package query

import (
	"context"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// GetTenantConfigQuery representa el caso de uso para obtener una configuración
type GetTenantConfigQuery struct {
	repository repository.TenantConfigRepository
}

// NewGetTenantConfigQuery crea una nueva instancia del query
func NewGetTenantConfigQuery(repo repository.TenantConfigRepository) *GetTenantConfigQuery {
	return &GetTenantConfigQuery{
		repository: repo,
	}
}

// Execute ejecuta el query
// Retorna (config, existe, error)
func (q *GetTenantConfigQuery) Execute(ctx context.Context, tenantID uuid.UUID, key string) (*entity.TenantConfig, bool, error) {
	return q.repository.GetByKey(ctx, tenantID, key)
}
