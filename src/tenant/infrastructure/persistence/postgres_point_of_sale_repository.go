package persistence

import (
	"context"
	"database/sql"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
	"github.com/hornosg/go-shared/infrastructure/postgres"
)

// PostgresPointOfSaleRepository implementa el repositorio usando PostgreSQL.
//
// RLS (E29, RULE-09/RULE-10): `points_of_sale` tiene ROW LEVEL SECURITY forzado con la
// policy `tenant_isolation` (migración 005). Cada operación corre dentro de
// postgres.WithRLSInTransaction, que fija `app.tenant_id` con SET LOCAL — sin él, cualquier
// query erra (fail-closed) bajo el rol NOBYPASSRLS `tenant_app`. El filtro manual
// `WHERE tenant_id = $` se mantiene como defensa en profundidad.
type PostgresPointOfSaleRepository struct {
	db *sql.DB
}

// NewPostgresPointOfSaleRepository crea una nueva instancia del repositorio
func NewPostgresPointOfSaleRepository(db *sql.DB) repository.PointOfSaleRepository {
	return &PostgresPointOfSaleRepository{db: db}
}

// Create crea un nuevo punto de venta. Tenant: pos.TenantID (path del WITH CHECK).
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

	rc := postgres.RLSContext{TenantID: pos.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query,
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
	})
}

// ListByTenant obtiene todos los puntos de venta de un tenant. Tenant: tenantID param.
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

	rc := postgres.RLSContext{TenantID: tenantID.String()}
	var pointsOfSale []*entity.PointOfSale
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var pos entity.PointOfSale
			if err := rows.Scan(
				&pos.ID,
				&pos.TenantID,
				&pos.Code,
				&pos.Description,
				&pos.IsFiscalEnabled,
				&pos.DefaultInvoiceType,
				&pos.IsActive,
				&pos.CreatedAt,
				&pos.Version,
			); err != nil {
				return err
			}
			pointsOfSale = append(pointsOfSale, &pos)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return pointsOfSale, nil
}

// ListActiveByTenant obtiene solo los puntos activos de un tenant. Tenant: tenantID param.
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

	rc := postgres.RLSContext{TenantID: tenantID.String()}
	var pointsOfSale []*entity.PointOfSale
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var pos entity.PointOfSale
			if err := rows.Scan(
				&pos.ID,
				&pos.TenantID,
				&pos.Code,
				&pos.Description,
				&pos.IsFiscalEnabled,
				&pos.DefaultInvoiceType,
				&pos.IsActive,
				&pos.CreatedAt,
				&pos.Version,
			); err != nil {
				return err
			}
			pointsOfSale = append(pointsOfSale, &pos)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return pointsOfSale, nil
}
