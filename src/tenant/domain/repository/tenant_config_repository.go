package repository

import (
	"context"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
)

// TenantConfigRepository define el contrato para el repositorio de configuraciones
type TenantConfigRepository interface {
	// GetByKey obtiene una configuración por tenant y clave
	// Retorna (config, existe, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*entity.TenantConfig, bool, error)

	// Save guarda o actualiza una configuración
	Save(ctx context.Context, config *entity.TenantConfig) error

	// Delete elimina una configuración
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error

	// GetAllByTenant obtiene todas las configuraciones de un tenant
	GetAllByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantConfig, error)
}
