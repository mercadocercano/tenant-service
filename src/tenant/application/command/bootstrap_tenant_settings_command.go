package command

import (
	"context"
	"log"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// BootstrapTenantSettingsCommand representa el caso de uso para inicializar settings por defecto
// Este command es idempotente: solo crea settings si no existen
type BootstrapTenantSettingsCommand struct {
	repository repository.TenantSettingsRepository
}

// NewBootstrapTenantSettingsCommand crea una nueva instancia del command
func NewBootstrapTenantSettingsCommand(repo repository.TenantSettingsRepository) *BootstrapTenantSettingsCommand {
	return &BootstrapTenantSettingsCommand{
		repository: repo,
	}
}

// Execute ejecuta el bootstrap de configuración estructurada
// Retorna (true si creó settings, error)
func (c *BootstrapTenantSettingsCommand) Execute(
	ctx context.Context,
	tenantID uuid.UUID,
	cashCustomerID uuid.UUID,
) (bool, error) {
	log.Printf("[BootstrapTenantSettings] Starting bootstrap for tenant %s", tenantID)

	// Verificar si ya existen settings
	exists, err := c.repository.Exists(ctx, tenantID)
	if err != nil {
		log.Printf("[BootstrapTenantSettings] Error checking if settings exist: %v", err)
		return false, err
	}

	if exists {
		log.Printf("[BootstrapTenantSettings] Settings already exist for tenant %s, skipping", tenantID)
		return false, nil
	}

	// Crear settings por defecto
	settings := entity.NewTenantSettings(tenantID, cashCustomerID)

	// Validar
	if err := settings.Validate(); err != nil {
		log.Printf("[BootstrapTenantSettings] Validation error: %v", err)
		return false, err
	}

	// Persistir
	if err := c.repository.Save(ctx, settings); err != nil {
		log.Printf("[BootstrapTenantSettings] Error saving settings: %v", err)
		return false, err
	}

	log.Printf("[BootstrapTenantSettings] Settings created successfully for tenant %s", tenantID)
	return true, nil
}
