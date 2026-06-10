package persistence

import (
	"context"
	"database/sql"
	"testing"
	"tenant/src/tenant/domain/entity"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestPostgresTenantConfigRepository_Save_Insert tests inserting a new config
func TestPostgresTenantConfigRepository_Save_Insert(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)

	tenantID := uuid.New()
	config := entity.NewTenantConfig(tenantID, "catalog.stock_policy", "IGNORE_STOCK")

	// Expect INSERT query
	mock.ExpectExec("INSERT INTO tenant_config").
		WithArgs(config.ID, config.TenantID, config.Key, config.Value, config.CreatedAt, config.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	err = repo.Save(context.Background(), config)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgresTenantConfigRepository_Save_Update tests updating an existing config
func TestPostgresTenantConfigRepository_Save_Update(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)

	tenantID := uuid.New()
	config := entity.NewTenantConfig(tenantID, "catalog.stock_policy", "IGNORE_STOCK")

	// Expect UPSERT query (INSERT ... ON CONFLICT DO UPDATE)
	mock.ExpectExec("INSERT INTO tenant_config").
		WithArgs(config.ID, config.TenantID, config.Key, config.Value, config.CreatedAt, config.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	err = repo.Save(context.Background(), config)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgresTenantConfigRepository_GetByKey_Found tests retrieving an existing config
func TestPostgresTenantConfigRepository_GetByKey_Found(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)

	tenantID := uuid.New()
	configID := uuid.New()
	key := "catalog.stock_policy"
	value := "IGNORE_STOCK"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "config_key", "config_value", "created_at", "updated_at"}).
		AddRow(configID, tenantID, key, value, now, now)

	mock.ExpectQuery("SELECT (.+) FROM tenant_config").
		WithArgs(tenantID, key).
		WillReturnRows(rows)

	// Act
	config, exists, err := repo.GetByKey(context.Background(), tenantID, key)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, config)
	assert.Equal(t, configID, config.ID)
	assert.Equal(t, tenantID, config.TenantID)
	assert.Equal(t, key, config.Key)
	assert.Equal(t, value, config.Value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgresTenantConfigRepository_GetByKey_NotFound tests retrieving a non-existing config
func TestPostgresTenantConfigRepository_GetByKey_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)

	tenantID := uuid.New()
	key := "catalog.stock_policy"

	mock.ExpectQuery("SELECT (.+) FROM tenant_config").
		WithArgs(tenantID, key).
		WillReturnError(sql.ErrNoRows)

	// Act
	config, exists, err := repo.GetByKey(context.Background(), tenantID, key)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, config)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantConfigRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)

	tenantID := uuid.New()
	key := "catalog.stock_policy"

	mock.ExpectExec("DELETE FROM tenant_config").
		WithArgs(tenantID, key).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), tenantID, key)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantConfigRepository_GetAllByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)

	tenantID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "config_key", "config_value", "created_at", "updated_at"}).
		AddRow(uuid.New(), tenantID, "catalog.stock_policy", "IGNORE_STOCK", now, now).
		AddRow(uuid.New(), tenantID, "fiscal.mode", "DISABLED", now, now)

	mock.ExpectQuery("SELECT (.+) FROM tenant_config").
		WithArgs(tenantID).
		WillReturnRows(rows)

	configs, err := repo.GetAllByTenant(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Len(t, configs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantConfigRepository_GetAllByTenant_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantConfigRepository(db)
	tenantID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM tenant_config").
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "config_key", "config_value", "created_at", "updated_at"}))

	configs, err := repo.GetAllByTenant(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Empty(t, configs)
	assert.NoError(t, mock.ExpectationsWereMet())
}
