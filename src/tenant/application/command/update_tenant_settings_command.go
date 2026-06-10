package command

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// EventPublisher define el contrato para publicar eventos
type EventPublisher interface {
	Publish(ctx context.Context, aggregateID, aggregateType, eventType string, payload []byte, publishedBy string) error
}

// UpdateTenantSettingsCommand representa el caso de uso para actualizar configuraciones
type UpdateTenantSettingsCommand struct {
	repository     repository.TenantSettingsRepository
	eventPublisher EventPublisher
}

// NewUpdateTenantSettingsCommand crea una nueva instancia del command
func NewUpdateTenantSettingsCommand(
	repo repository.TenantSettingsRepository,
	eventPublisher EventPublisher,
) *UpdateTenantSettingsCommand {
	return &UpdateTenantSettingsCommand{
		repository:     repo,
		eventPublisher: eventPublisher,
	}
}

// UpdateTenantSettingsParams contiene todos los parámetros necesarios
type UpdateTenantSettingsParams struct {
	TenantID                         uuid.UUID
	Version                          int
	BaseCurrency                     string
	AllowedCurrencies                []string
	ExchangeRateSource               string
	AutoUpdateExchangeRate           bool
	FiscalMode                       string
	InvoiceGeneration                string
	AllowSaleIfAfipFails             bool
	AutoRetryFailedInvoices          bool
	EmailInvoiceAfterSuccess         bool
	DefaultInvoiceType               string
	TaxRegime                        string
	StockPolicy                      string
	AllowNegativeStock               bool
	RequireStockValidationBeforeSale bool
	CreditEnabled                    bool
	DefaultCreditDays                int
	MaxCreditLimit                   float64
	AllowSaleOverCreditLimit         bool
	CashCustomerID                   uuid.UUID
}

// Execute ejecuta el command con optimistic locking y publicación de evento
func (c *UpdateTenantSettingsCommand) Execute(
	ctx context.Context,
	params UpdateTenantSettingsParams,
) (*entity.TenantSettings, error) {
	log.Printf("[UpdateTenantSettings] Starting update for tenant %s with version %d", params.TenantID, params.Version)

	// 1. Cargar configuración actual
	currentSettings, err := c.repository.GetByTenantID(ctx, params.TenantID)
	if err != nil {
		log.Printf("[UpdateTenantSettings] Error loading current settings: %v", err)
		return nil, err
	}

	// 2. Validar version (optimistic locking)
	if currentSettings.Version != params.Version {
		log.Printf("[UpdateTenantSettings] Version conflict: expected %d, got %d", params.Version, currentSettings.Version)
		return nil, errors.New("version conflict: settings were modified by another transaction")
	}

	// 3. Actualizar campos
	currentSettings.Update(
		params.BaseCurrency,
		params.AllowedCurrencies,
		params.ExchangeRateSource,
		params.AutoUpdateExchangeRate,
		params.FiscalMode,
		params.InvoiceGeneration,
		params.AllowSaleIfAfipFails,
		params.AutoRetryFailedInvoices,
		params.EmailInvoiceAfterSuccess,
		params.DefaultInvoiceType,
		params.TaxRegime,
		params.StockPolicy,
		params.AllowNegativeStock,
		params.RequireStockValidationBeforeSale,
		params.CreditEnabled,
		params.DefaultCreditDays,
		params.MaxCreditLimit,
		params.AllowSaleOverCreditLimit,
		params.CashCustomerID,
	)

	// 4. Validar reglas de negocio
	if err := currentSettings.Validate(); err != nil {
		log.Printf("[UpdateTenantSettings] Validation error: %v", err)
		return nil, err
	}

	// 5. Incrementar versión
	currentSettings.IncrementVersion()

	// 6. Persistir cambios
	if err := c.repository.Save(ctx, currentSettings); err != nil {
		log.Printf("[UpdateTenantSettings] Error saving settings: %v", err)
		return nil, err
	}

	log.Printf("[UpdateTenantSettings] Settings updated successfully. New version: %d", currentSettings.Version)

	// 7. Publicar evento tenant.settings.updated
	if err := c.publishSettingsUpdatedEvent(ctx, currentSettings); err != nil {
		// Log el error pero no fallar la transacción
		// El evento será publicado eventualmente por el worker
		log.Printf("[UpdateTenantSettings] Warning: Failed to publish event: %v", err)
	} else {
		log.Printf("[UpdateTenantSettings] Event published successfully")
	}

	return currentSettings, nil
}

// publishSettingsUpdatedEvent publica el evento tenant.settings.updated
func (c *UpdateTenantSettingsCommand) publishSettingsUpdatedEvent(
	ctx context.Context,
	settings *entity.TenantSettings,
) error {
	// Crear payload con snapshot completo de configuraciones relevantes
	payload := map[string]interface{}{
		"version":                              settings.Version,
		"base_currency":                        settings.BaseCurrency,
		"allowed_currencies":                   settings.AllowedCurrencies,
		"fiscal_mode":                          settings.FiscalMode,
		"invoice_generation":                   settings.InvoiceGeneration,
		"stock_policy":                         settings.StockPolicy,
		"allow_negative_stock":                 settings.AllowNegativeStock,
		"credit_enabled":                       settings.CreditEnabled,
		"max_credit_limit":                     settings.MaxCreditLimit,
		"default_credit_days":                  settings.DefaultCreditDays,
		"cash_customer_id":                     settings.CashCustomerID.String(),
		"tax_regime":                           settings.TaxRegime,
		"exchange_rate_source":                 settings.ExchangeRateSource,
		"require_stock_validation_before_sale": settings.RequireStockValidationBeforeSale,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.eventPublisher.Publish(
		ctx,
		settings.TenantID.String(), // aggregate_id
		"tenant_settings",          // aggregate_type
		"tenant.settings.updated",  // event_type
		payloadBytes,
		"tenant-service", // published_by
	)
}
