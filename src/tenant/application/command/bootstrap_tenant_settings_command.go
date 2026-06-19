package command

import (
	"context"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/port"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// BootstrapTenantSettingsCommand representa el caso de uso para inicializar settings por defecto
// Este command es idempotente: solo crea settings si no existen
type BootstrapTenantSettingsCommand struct {
	repository repository.TenantSettingsRepository
	logger     port.TenantEventLogger
}

// NewBootstrapTenantSettingsCommand crea una nueva instancia del command
func NewBootstrapTenantSettingsCommand(repo repository.TenantSettingsRepository) *BootstrapTenantSettingsCommand {
	return &BootstrapTenantSettingsCommand{
		repository: repo,
	}
}

// NewBootstrapTenantSettingsCommandWithLogger crea una instancia con logger canónico inyectado.
func NewBootstrapTenantSettingsCommandWithLogger(repo repository.TenantSettingsRepository, logger port.TenantEventLogger) *BootstrapTenantSettingsCommand {
	return &BootstrapTenantSettingsCommand{
		repository: repo,
		logger:     logger,
	}
}

// logEvent emite un evento canónico si hay logger inyectado (nil-safe).
func (c *BootstrapTenantSettingsCommand) logEvent(e port.TenantEvent) {
	if c.logger != nil {
		c.logger.Log(e)
	}
}

// Execute ejecuta el bootstrap de configuración estructurada
// Retorna (true si creó settings, error)
func (c *BootstrapTenantSettingsCommand) Execute(
	ctx context.Context,
	tenantID uuid.UUID,
	cashCustomerID uuid.UUID,
) (bool, error) {
	// Verificar si ya existen settings
	exists, err := c.repository.Exists(ctx, tenantID)
	if err != nil {
		c.logEvent(port.TenantEvent{Event: "tenant.settings_bootstrap_failed", TenantID: tenantID.String(), Reason: err.Error()})
		return false, err
	}

	if exists {
		c.logEvent(port.TenantEvent{Event: "tenant.settings_already_exists", TenantID: tenantID.String()})
		return false, nil
	}

	// Crear settings por defecto
	settings := entity.NewTenantSettings(tenantID, cashCustomerID)

	// Validar
	if err := settings.Validate(); err != nil {
		c.logEvent(port.TenantEvent{Event: "tenant.settings_bootstrap_failed", TenantID: tenantID.String(), Reason: err.Error()})
		return false, err
	}

	// Persistir
	if err := c.repository.Save(ctx, settings); err != nil {
		c.logEvent(port.TenantEvent{Event: "tenant.settings_bootstrap_failed", TenantID: tenantID.String(), Reason: err.Error()})
		return false, err
	}

	c.logEvent(port.TenantEvent{Event: "tenant.settings_bootstrapped", TenantID: tenantID.String()})
	return true, nil
}
