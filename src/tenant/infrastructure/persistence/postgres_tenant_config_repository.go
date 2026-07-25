package persistence

import (
	"context"
	"database/sql"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
	"github.com/hornosg/go-shared/infrastructure/postgres"
)

// PostgresTenantConfigRepository implementa el repositorio usando PostgreSQL.
//
// RLS (E29, RULE-09/RULE-10): `tenant_config` tiene ROW LEVEL SECURITY forzado con la
// policy `tenant_isolation` (migración 005). Cada operación corre dentro de
// postgres.WithRLSInTransaction, que fija `app.tenant_id` con SET LOCAL — sin él, cualquier
// query erra (fail-closed) bajo el rol NOBYPASSRLS `tenant_app`. El filtro manual
// `WHERE tenant_id = $` se mantiene como defensa en profundidad.
type PostgresTenantConfigRepository struct {
	db *sql.DB
}

// NewPostgresTenantConfigRepository crea una nueva instancia del repositorio
func NewPostgresTenantConfigRepository(db *sql.DB) repository.TenantConfigRepository {
	return &PostgresTenantConfigRepository{db: db}
}

// GetByKey obtiene una configuración por tenant y clave. Tenant: tenantID param.
func (r *PostgresTenantConfigRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*entity.TenantConfig, bool, error) {
	query := `
		SELECT id, tenant_id, config_key, config_value, created_at, updated_at
		FROM tenant_config
		WHERE tenant_id = $1 AND config_key = $2
	`

	rc := postgres.RLSContext{TenantID: tenantID.String()}
	var config entity.TenantConfig
	var found bool
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx, query, tenantID, key).Scan(
			&config.ID,
			&config.TenantID,
			&config.Key,
			&config.Value,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		found = true
		return nil
	})

	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	return &config, true, nil
}

// Save guarda o actualiza una configuración (upsert). Tenant: config.TenantID (path del
// WITH CHECK).
func (r *PostgresTenantConfigRepository) Save(ctx context.Context, config *entity.TenantConfig) error {
	query := `
		INSERT INTO tenant_config (id, tenant_id, config_key, config_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, config_key)
		DO UPDATE SET
			config_value = EXCLUDED.config_value,
			updated_at = EXCLUDED.updated_at
	`

	rc := postgres.RLSContext{TenantID: config.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query,
			config.ID,
			config.TenantID,
			config.Key,
			config.Value,
			config.CreatedAt,
			config.UpdatedAt,
		)
		return err
	})
}

