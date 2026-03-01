package command

import (
	"context"
	"log"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// BootstrapTenantConfigCommand representa el caso de uso para inicializar configuración default
// Este command es idempotente: solo crea config si no existe
type BootstrapTenantConfigCommand struct {
	repository repository.TenantConfigRepository
}

// NewBootstrapTenantConfigCommand crea una nueva instancia del command
func NewBootstrapTenantConfigCommand(repo repository.TenantConfigRepository) *BootstrapTenantConfigCommand {
	return &BootstrapTenantConfigCommand{
		repository: repo,
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
	log.Printf("=== BOOTSTRAP TENANT CONFIG START ===")
	log.Printf("TenantID: %s", tenantID)

	createdCount := 0

	// Iterar sobre las configuraciones default
	for key, value := range DefaultConfigs {
		log.Printf("Checking config: %s", key)

		// Verificar si ya existe
		_, exists, err := c.repository.GetByKey(ctx, tenantID, key)
		if err != nil {
			log.Printf("ERROR: Failed to check existing config for key %s: %v", key, err)
			return createdCount, err
		}

		if exists {
			log.Printf("Config %s already exists, skipping", key)
			continue
		}

		// No existe, crear nueva configuración
		log.Printf("Creating default config: %s = %s", key, value)
		config := entity.NewTenantConfig(tenantID, key, value)

		if err := c.repository.Save(ctx, config); err != nil {
			log.Printf("ERROR: Failed to save config %s: %v", key, err)
			return createdCount, err
		}

		createdCount++
		log.Printf("Config %s created successfully", key)
	}

	log.Printf("Bootstrap completed: %d configs created", createdCount)
	log.Printf("=== BOOTSTRAP TENANT CONFIG END ===")

	return createdCount, nil
}
