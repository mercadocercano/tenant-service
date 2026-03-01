package command

import (
	"context"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// SetTenantConfigCommand representa el caso de uso para establecer una configuración
type SetTenantConfigCommand struct {
	repository repository.TenantConfigRepository
}

// NewSetTenantConfigCommand crea una nueva instancia del command
func NewSetTenantConfigCommand(repo repository.TenantConfigRepository) *SetTenantConfigCommand {
	return &SetTenantConfigCommand{
		repository: repo,
	}
}

// Execute ejecuta el command (upsert)
// Retorna la configuración guardada
func (c *SetTenantConfigCommand) Execute(ctx context.Context, tenantID uuid.UUID, key, value string) (*entity.TenantConfig, error) {
	// Buscar si ya existe
	existingConfig, exists, err := c.repository.GetByKey(ctx, tenantID, key)
	if err != nil {
		return nil, err
	}

	var config *entity.TenantConfig

	if exists {
		// UPDATE: actualizar el valor existente
		existingConfig.Update(value)
		config = existingConfig
	} else {
		// INSERT: crear nueva configuración
		config = entity.NewTenantConfig(tenantID, key, value)
	}

	// Guardar (upsert)
	if err := c.repository.Save(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}
