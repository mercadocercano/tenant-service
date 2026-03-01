package repository

import (
	"context"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
)

// TenantSettingsRepository define el contrato para persistir configuraciones estructuradas
type TenantSettingsRepository interface {
	// GetByTenantID obtiene la configuración de un tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entity.TenantSettings, error)

	// Save guarda o actualiza la configuración (upsert con optimistic locking)
	// Retorna error si la versión no coincide (version conflict)
	Save(ctx context.Context, settings *entity.TenantSettings) error

	// Exists verifica si existe configuración para un tenant
	Exists(ctx context.Context, tenantID uuid.UUID) (bool, error)
}
