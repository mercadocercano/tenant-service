package command

import (
	"context"
	"encoding/json"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/port"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// CreatePointOfSaleCommand representa el caso de uso para crear un punto de venta
type CreatePointOfSaleCommand struct {
	repository     repository.PointOfSaleRepository
	eventPublisher EventPublisher
	logger         port.TenantEventLogger
}

// NewCreatePointOfSaleCommand crea una nueva instancia del command
func NewCreatePointOfSaleCommand(repo repository.PointOfSaleRepository, eventPublisher EventPublisher) *CreatePointOfSaleCommand {
	return &CreatePointOfSaleCommand{
		repository:     repo,
		eventPublisher: eventPublisher,
	}
}

// NewCreatePointOfSaleCommandWithLogger crea una instancia con logger canónico inyectado.
func NewCreatePointOfSaleCommandWithLogger(repo repository.PointOfSaleRepository, eventPublisher EventPublisher, logger port.TenantEventLogger) *CreatePointOfSaleCommand {
	return &CreatePointOfSaleCommand{
		repository:     repo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// logEvent emite un evento canónico si hay logger inyectado (nil-safe).
func (c *CreatePointOfSaleCommand) logEvent(e port.TenantEvent) {
	if c.logger != nil {
		c.logger.Log(e)
	}
}

// Execute ejecuta el command
func (c *CreatePointOfSaleCommand) Execute(
	ctx context.Context,
	tenantID uuid.UUID,
	code int,
	description string,
	isFiscalEnabled bool,
	defaultInvoiceType string,
) (*entity.PointOfSale, error) {
	// 1. Crear entidad
	pos := entity.NewPointOfSale(tenantID, code, description, isFiscalEnabled, defaultInvoiceType)

	// 2. Validar
	if err := pos.Validate(); err != nil {
		c.logEvent(port.TenantEvent{Event: "tenant.pos_create_failed", TenantID: tenantID.String(), Reason: err.Error()})
		return nil, err
	}

	// 3. Persistir
	if err := c.repository.Create(ctx, pos); err != nil {
		c.logEvent(port.TenantEvent{Event: "tenant.pos_create_failed", TenantID: tenantID.String(), Reason: err.Error()})
		return nil, err
	}

	c.logEvent(port.TenantEvent{Event: "tenant.pos_created", TenantID: tenantID.String(), PosID: pos.ID.String()})

	// 4. Publicar evento tenant.point_of_sale.created
	if err := c.publishPOSCreatedEvent(ctx, pos); err != nil {
		_ = err // intencional: error no falla la operación (venta ya registrada)
	}

	return pos, nil
}

func (c *CreatePointOfSaleCommand) publishPOSCreatedEvent(ctx context.Context, pos *entity.PointOfSale) error {
	payload := map[string]interface{}{
		"id":                   pos.ID.String(),
		"tenant_id":            pos.TenantID.String(),
		"code":                 pos.Code,
		"description":          pos.Description,
		"is_fiscal_enabled":    pos.IsFiscalEnabled,
		"default_invoice_type": pos.DefaultInvoiceType,
		"is_active":            pos.IsActive,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.eventPublisher.Publish(
		ctx,
		pos.ID.String(),
		"point_of_sale",
		"tenant.point_of_sale.created",
		payloadBytes,
		"tenant-service",
	)
}
