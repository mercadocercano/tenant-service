package config

import (
	"database/sql"
	"tenant/src/tenant/application/command"
	"tenant/src/tenant/application/query"
	"tenant/src/tenant/domain/port"
	"tenant/src/tenant/infrastructure/controller"
	"tenant/src/tenant/infrastructure/persistence"
)

// TenantModuleConfig contiene las dependencias del módulo tenant
type TenantModuleConfig struct {
	ConfigController      *controller.TenantConfigController
	SettingsController    *controller.TenantSettingsController
	PointOfSaleController *controller.PointOfSaleController
}

// NewTenantModuleConfig crea e inicializa el módulo tenant (key-value existente)
func NewTenantModuleConfig(db *sql.DB) *TenantModuleConfig {
	// Crear repositorio
	repo := persistence.NewPostgresTenantConfigRepository(db)

	// Crear queries
	getConfigQuery := query.NewGetTenantConfigQuery(repo)

	// Crear commands
	setConfigCommand := command.NewSetTenantConfigCommand(repo)
	bootstrapCommand := command.NewBootstrapTenantConfigCommand(repo)

	// Crear controlador
	ctrl := controller.NewTenantConfigController(getConfigQuery, setConfigCommand, bootstrapCommand)

	return &TenantModuleConfig{
		ConfigController: ctrl,
	}
}

// NewExtendedTenantModuleConfig crea el módulo completo con configuraciones estructuradas
func NewExtendedTenantModuleConfig(db *sql.DB, eventPublisher command.EventPublisher) *TenantModuleConfig {
	return NewExtendedTenantModuleConfigWithLogger(db, eventPublisher, nil)
}

// NewExtendedTenantModuleConfigWithLogger crea el módulo completo inyectando un logger canónico (ADR-001).
func NewExtendedTenantModuleConfigWithLogger(db *sql.DB, eventPublisher command.EventPublisher, logger port.TenantEventLogger) *TenantModuleConfig {
	// === KEY-VALUE (EXISTENTE) ===
	configRepo := persistence.NewPostgresTenantConfigRepository(db)
	getConfigQuery := query.NewGetTenantConfigQuery(configRepo)
	setConfigCommand := command.NewSetTenantConfigCommand(configRepo)
	bootstrapCommand := command.NewBootstrapTenantConfigCommandWithLogger(configRepo, logger)
	configController := controller.NewTenantConfigController(getConfigQuery, setConfigCommand, bootstrapCommand)

	// === TENANT SETTINGS (NUEVO) ===
	settingsRepo := persistence.NewPostgresTenantSettingsRepository(db)
	getSettingsQuery := query.NewGetTenantSettingsQuery(settingsRepo)
	updateSettingsCommand := command.NewUpdateTenantSettingsCommandWithLogger(settingsRepo, eventPublisher, logger)
	settingsController := controller.NewTenantSettingsController(getSettingsQuery, updateSettingsCommand)

	// === POINTS OF SALE (NUEVO) ===
	posRepo := persistence.NewPostgresPointOfSaleRepository(db)
	createPOSCommand := command.NewCreatePointOfSaleCommandWithLogger(posRepo, eventPublisher, logger)
	listPOSQuery := query.NewListPointsOfSaleQuery(posRepo)
	posController := controller.NewPointOfSaleController(createPOSCommand, listPOSQuery)

	return &TenantModuleConfig{
		ConfigController:      configController,
		SettingsController:    settingsController,
		PointOfSaleController: posController,
	}
}
