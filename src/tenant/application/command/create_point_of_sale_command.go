package command

import (
	"context"
	"log"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// CreatePointOfSaleCommand representa el caso de uso para crear un punto de venta
type CreatePointOfSaleCommand struct {
	repository repository.PointOfSaleRepository
}

// NewCreatePointOfSaleCommand crea una nueva instancia del command
func NewCreatePointOfSaleCommand(repo repository.PointOfSaleRepository) *CreatePointOfSaleCommand {
	return &CreatePointOfSaleCommand{
		repository: repo,
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

	// TODO: Publicar evento tenant.point_of_sale.created si es necesario

	return pos, nil
}
