package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tenant/src/tenant/domain/entity"
)

func newSettingsWithVersion(tenantID uuid.UUID, version int) *entity.TenantSettings {
	s := entity.NewTenantSettings(tenantID, uuid.New())
	s.Version = version
	s.UpdatedAt = time.Now()
	return s
}

func TestPostgresTenantSettingsRepository_GetByTenantID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)

	tenantID := uuid.New()
	cashID := uuid.New()
	currencies := []string{"ARS", "USD"}
	currJSON, _ := json.Marshal(currencies)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"tenant_id", "base_currency", "allowed_currencies", "exchange_rate_source",
		"auto_update_exchange_rate", "fiscal_mode", "invoice_generation",
		"allow_sale_if_afip_fails", "auto_retry_failed_invoices", "email_invoice_after_success",
		"default_invoice_type", "tax_regime", "stock_policy", "allow_negative_stock",
		"require_stock_validation_before_sale", "credit_enabled", "default_credit_days",
		"max_credit_limit", "allow_sale_over_credit_limit", "cash_customer_id", "version", "updated_at",
	}).AddRow(
		tenantID, "ARS", currJSON, "MANUAL",
		false, "DISABLED", "MANUAL",
		true, false, false,
		"B", "MONOTRIBUTO", "IGNORE", true,
		false, false, 30,
		0.0, false, cashID, 1, now,
	)

	mock.ExpectQuery("SELECT").WithArgs(tenantID).WillReturnRows(rows)

	result, err := repo.GetByTenantID(context.Background(), tenantID)

	require.NoError(t, err)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, "ARS", result.BaseCurrency)
	assert.Equal(t, currencies, result.AllowedCurrencies)
	assert.Equal(t, 1, result.Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_GetByTenantID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)
	tenantID := uuid.New()

	mock.ExpectQuery("SELECT").WithArgs(tenantID).WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByTenantID(context.Background(), tenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_Save_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)

	tenantID := uuid.New()
	settings := newSettingsWithVersion(tenantID, 2) // version=2, previousVersion=1

	mock.ExpectExec("UPDATE tenant_settings").WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Save(context.Background(), settings)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_Save_Update_VersionConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)

	tenantID := uuid.New()
	settings := newSettingsWithVersion(tenantID, 2)

	// UPDATE afecta 0 filas (version conflict o no existe)
	mock.ExpectExec("UPDATE tenant_settings").WillReturnResult(sqlmock.NewResult(0, 0))
	// Exists retorna true → version conflict
	mock.ExpectQuery("SELECT EXISTS").WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	err = repo.Save(context.Background(), settings)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version conflict")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_Save_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)

	tenantID := uuid.New()
	settings := newSettingsWithVersion(tenantID, 1) // version=1, previousVersion=0

	// UPDATE afecta 0 filas (no existe)
	mock.ExpectExec("UPDATE tenant_settings").WillReturnResult(sqlmock.NewResult(0, 0))
	// Exists retorna false → INSERT
	mock.ExpectQuery("SELECT EXISTS").WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// INSERT
	mock.ExpectExec("INSERT INTO tenant_settings").WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(context.Background(), settings)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_Exists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)
	tenantID := uuid.New()

	mock.ExpectQuery("SELECT EXISTS").WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_Exists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)
	tenantID := uuid.New()

	mock.ExpectQuery("SELECT EXISTS").WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.Exists(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTenantSettingsRepository_Save_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantSettingsRepository(db)

	tenantID := uuid.New()
	settings := newSettingsWithVersion(tenantID, 2)

	mock.ExpectExec("UPDATE tenant_settings").WillReturnError(errors.New("connection reset"))

	err = repo.Save(context.Background(), settings)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection reset")
	assert.NoError(t, mock.ExpectationsWereMet())
}
