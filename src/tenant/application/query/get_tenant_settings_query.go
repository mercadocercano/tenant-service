package query

import (
	"context"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// GetTenantSettingsQuery representa el caso de uso para obtener configuraciones
type GetTenantSettingsQuery struct {
	repository repository.TenantSettingsRepository
}

// NewGetTenantSettingsQuery crea una nueva instancia del query
func NewGetTenantSettingsQuery(repo repository.TenantSettingsRepository) *GetTenantSettingsQuery {
	return &GetTenantSettingsQuery{
		repository: repo,
	}
}

// Execute ejecuta el query
func (q *GetTenantSettingsQuery) Execute(ctx context.Context, tenantID uuid.UUID) (*entity.TenantSettings, error) {
	return q.repository.GetByTenantID(ctx, tenantID)
}
