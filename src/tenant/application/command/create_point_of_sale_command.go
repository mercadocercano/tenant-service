package command

import (
	"context"
	"encoding/json"
	"log"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// CreatePointOfSaleCommand representa el caso de uso para crear un punto de venta
type CreatePointOfSaleCommand struct {
	repository     repository.PointOfSaleRepository
	eventPublisher EventPublisher
}

// NewCreatePointOfSaleCommand crea una nueva instancia del command
func NewCreatePointOfSaleCommand(repo repository.PointOfSaleRepository, eventPublisher EventPublisher) *CreatePointOfSaleCommand {
	return &CreatePointOfSaleCommand{
		repository:     repo,
		eventPublisher: eventPublisher,
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
	log.Printf("[CreatePointOfSale] Creating POS for tenant %s, code %d", tenantID, code)

	// 1. Crear entidad
	pos := entity.NewPointOfSale(tenantID, code, description, isFiscalEnabled, defaultInvoiceType)

	// 2. Validar
	if err := pos.Validate(); err != nil {
		log.Printf("[CreatePointOfSale] Validation error: %v", err)
		return nil, err
	}

	// 3. Persistir
	if err := c.repository.Create(ctx, pos); err != nil {
		log.Printf("[CreatePointOfSale] Error creating POS: %v", err)
		return nil, err
	}

	log.Printf("[CreatePointOfSale] POS created successfully: %s", pos.ID)

	// 4. Publicar evento tenant.point_of_sale.created
	if err := c.publishPOSCreatedEvent(ctx, pos); err != nil {
		log.Printf("[CreatePointOfSale] WARNING: failed to publish event: %v", err)
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
