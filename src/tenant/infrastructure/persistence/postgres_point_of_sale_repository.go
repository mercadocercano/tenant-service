package persistence

import (
	"context"
	"database/sql"
	"errors"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// PostgresPointOfSaleRepository implementa el repositorio usando PostgreSQL
type PostgresPointOfSaleRepository struct {
	db *sql.DB
}

// NewPostgresPointOfSaleRepository crea una nueva instancia del repositorio
func NewPostgresPointOfSaleRepository(db *sql.DB) repository.PointOfSaleRepository {
	return &PostgresPointOfSaleRepository{db: db}
}

// Create crea un nuevo punto de venta
func (r *PostgresPointOfSaleRepository) Create(ctx context.Context, pos *entity.PointOfSale) error {
	query := `
		INSERT INTO points_of_sale (
			id,
			tenant_id,
			code,
			description,
			is_fiscal_enabled,
			default_invoice_type,
			is_active,
			created_at,
			version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		pos.ID,
		pos.TenantID,
		pos.Code,
		pos.Description,
		pos.IsFiscalEnabled,
		pos.DefaultInvoiceType,
		pos.IsActive,
		pos.CreatedAt,
		pos.Version,
	)

	return err
}

// GetByID obtiene un punto de venta por ID
func (r *PostgresPointOfSaleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PointOfSale, error) {
	query := `
		SELECT 
			id,
			tenant_id,
			code,
			description,
			is_fiscal_enabled,
			default_invoice_type,
			is_active,
			created_at,
			version
		FROM points_of_sale
		WHERE id = $1
	`

	var pos entity.PointOfSale
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pos.ID,
		&pos.TenantID,
		&pos.Code,
		&pos.Description,
		&pos.IsFiscalEnabled,
		&pos.DefaultInvoiceType,
		&pos.IsActive,
		&pos.CreatedAt,
		&pos.Version,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("point of sale not found")
	}

	if err != nil {
		return nil, err
	}

	return &pos, nil
}

// ListByTenant obtiene todos los puntos de venta de un tenant
func (r *PostgresPointOfSaleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error) {
	query := `
		SELECT 
			id,
			tenant_id,
			code,
			description,
			is_fiscal_enabled,
			default_invoice_type,
			is_active,
			created_at,
			version
		FROM points_of_sale
		WHERE tenant_id = $1
		ORDER BY code ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pointsOfSale []*entity.PointOfSale
	for rows.Next() {
		var pos entity.PointOfSale
		err := rows.Scan(
			&pos.ID,
			&pos.TenantID,
			&pos.Code,
			&pos.Description,
			&pos.IsFiscalEnabled,
			&pos.DefaultInvoiceType,
			&pos.IsActive,
			&pos.CreatedAt,
			&pos.Version,
		)
		if err != nil {
			return nil, err
		}
		pointsOfSale = append(pointsOfSale, &pos)
	}

	return pointsOfSale, rows.Err()
}

// ListActiveByTenant obtiene solo los puntos activos de un tenant
func (r *PostgresPointOfSaleRepository) ListActiveByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error) {
	query := `
		SELECT 
			id,
			tenant_id,
			code,
			description,
			is_fiscal_enabled,
			default_invoice_type,
			is_active,
			created_at,
			version
		FROM points_of_sale
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY code ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pointsOfSale []*entity.PointOfSale
	for rows.Next() {
		var pos entity.PointOfSale
		err := rows.Scan(
			&pos.ID,
			&pos.TenantID,
			&pos.Code,
			&pos.Description,
			&pos.IsFiscalEnabled,
			&pos.DefaultInvoiceType,
			&pos.IsActive,
			&pos.CreatedAt,
			&pos.Version,
		)
		if err != nil {
			return nil, err
		}
		pointsOfSale = append(pointsOfSale, &pos)
	}

	return pointsOfSale, rows.Err()
}

// Update actualiza un punto de venta existente
func (r *PostgresPointOfSaleRepository) Update(ctx context.Context, pos *entity.PointOfSale) error {
	query := `
		UPDATE points_of_sale
		SET 
			description = $2,
			is_fiscal_enabled = $3,
			default_invoice_type = $4,
			is_active = $5,
			version = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		pos.ID,
		pos.Description,
		pos.IsFiscalEnabled,
		pos.DefaultInvoiceType,
		pos.IsActive,
		pos.Version,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("point of sale not found")
	}

	return nil
}
