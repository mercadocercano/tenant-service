package command

import (
	"context"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/port"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// BootstrapTenantConfigCommand representa el caso de uso para inicializar configuración default
// Este command es idempotente: solo crea config si no existe
type BootstrapTenantConfigCommand struct {
	repository repository.TenantConfigRepository
	logger     port.TenantEventLogger
}

// NewBootstrapTenantConfigCommand crea una nueva instancia del command
func NewBootstrapTenantConfigCommand(repo repository.TenantConfigRepository) *BootstrapTenantConfigCommand {
	return &BootstrapTenantConfigCommand{
		repository: repo,
	}
}

// NewBootstrapTenantConfigCommandWithLogger crea una instancia con logger canónico inyectado.
func NewBootstrapTenantConfigCommandWithLogger(repo repository.TenantConfigRepository, logger port.TenantEventLogger) *BootstrapTenantConfigCommand {
	return &BootstrapTenantConfigCommand{
		repository: repo,
		logger:     logger,
	}
}

// logEvent emite un evento canónico si hay logger inyectado (nil-safe).
func (c *BootstrapTenantConfigCommand) logEvent(e port.TenantEvent) {
	if c.logger != nil {
		c.logger.Log(e)
	}
}

// DefaultConfigs define las configuraciones por defecto para un tenant nuevo
var DefaultConfigs = map[string]string{
	"catalog.stock_policy": "REQUIRE_STOCK",
}

// Execute ejecuta el bootstrap de configuración
// Retorna (configs_creadas, error)
// Es idempotente: si ya existen configs, no hace nada
func (c *BootstrapTenantConfigCommand) Execute(ctx context.Context, tenantID uuid.UUID) (int, error) {
	createdCount := 0

	// Iterar sobre las configuraciones default
	for key, value := range DefaultConfigs {
		// Verificar si ya existe
		_, exists, err := c.repository.GetByKey(ctx, tenantID, key)
		if err != nil {
			c.logEvent(port.TenantEvent{Event: "tenant.config_bootstrap_failed", TenantID: tenantID.String(), ConfigKey: key, Reason: err.Error()})
			return createdCount, err
		}

		if exists {
			c.logEvent(port.TenantEvent{Event: "tenant.config_already_exists", TenantID: tenantID.String(), ConfigKey: key})
			continue
		}

		// No existe, crear nueva configuración
		config := entity.NewTenantConfig(tenantID, key, value)

		if err := c.repository.Save(ctx, config); err != nil {
			c.logEvent(port.TenantEvent{Event: "tenant.config_bootstrap_failed", TenantID: tenantID.String(), ConfigKey: key, Reason: err.Error()})
			return createdCount, err
		}

		createdCount++
	}

	c.logEvent(port.TenantEvent{Event: "tenant.config_bootstrapped", TenantID: tenantID.String(), ConfigsCreated: createdCount})
	return createdCount, nil
}
