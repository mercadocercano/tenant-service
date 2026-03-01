package persistence

import (
	"context"
	"database/sql"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
)

// PostgresTenantConfigRepository implementa el repositorio usando PostgreSQL
type PostgresTenantConfigRepository struct {
	db *sql.DB
}

// NewPostgresTenantConfigRepository crea una nueva instancia del repositorio
func NewPostgresTenantConfigRepository(db *sql.DB) repository.TenantConfigRepository {
	return &PostgresTenantConfigRepository{db: db}
}

// GetByKey obtiene una configuración por tenant y clave
func (r *PostgresTenantConfigRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*entity.TenantConfig, bool, error) {
	query := `
		SELECT id, tenant_id, config_key, config_value, created_at, updated_at
		FROM tenant_config
		WHERE tenant_id = $1 AND config_key = $2
	`

	var config entity.TenantConfig
	err := r.db.QueryRowContext(ctx, query, tenantID, key).Scan(
		&config.ID,
		&config.TenantID,
		&config.Key,
		&config.Value,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return &config, true, nil
}

// Save guarda o actualiza una configuración
func (r *PostgresTenantConfigRepository) Save(ctx context.Context, config *entity.TenantConfig) error {
	query := `
		INSERT INTO tenant_config (id, tenant_id, config_key, config_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, config_key)
		DO UPDATE SET
			config_value = EXCLUDED.config_value,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID,
		config.TenantID,
		config.Key,
		config.Value,
		config.CreatedAt,
		config.UpdatedAt,
	)

	return err
}

// Delete elimina una configuración
func (r *PostgresTenantConfigRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	query := `
		DELETE FROM tenant_config
		WHERE tenant_id = $1 AND config_key = $2
	`

	_, err := r.db.ExecContext(ctx, query, tenantID, key)
	return err
}

// GetAllByTenant obtiene todas las configuraciones de un tenant
func (r *PostgresTenantConfigRepository) GetAllByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantConfig, error) {
	query := `
		SELECT id, tenant_id, config_key, config_value, created_at, updated_at
		FROM tenant_config
		WHERE tenant_id = $1
		ORDER BY config_key
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*entity.TenantConfig
	for rows.Next() {
		var config entity.TenantConfig
		err := rows.Scan(
			&config.ID,
			&config.TenantID,
			&config.Key,
			&config.Value,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, &config)
	}

	return configs, rows.Err()
}
