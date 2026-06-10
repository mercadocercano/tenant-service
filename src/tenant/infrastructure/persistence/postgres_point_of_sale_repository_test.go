package persistence

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tenant/src/tenant/domain/entity"
)

func posColumns() []string {
	return []string{"id", "tenant_id", "code", "description", "is_fiscal_enabled",
		"default_invoice_type", "is_active", "created_at", "version"}
}

func addPOSRow(rows *sqlmock.Rows, pos *entity.PointOfSale) *sqlmock.Rows {
	return rows.AddRow(pos.ID, pos.TenantID, pos.Code, pos.Description,
		pos.IsFiscalEnabled, pos.DefaultInvoiceType, pos.IsActive, pos.CreatedAt, pos.Version)
}

func TestPostgresPointOfSaleRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)

	pos := entity.NewPointOfSale(uuid.New(), 1, "Sucursal Central", true, "B")

	mock.ExpectExec("INSERT INTO points_of_sale").
		WithArgs(pos.ID, pos.TenantID, pos.Code, pos.Description,
			pos.IsFiscalEnabled, pos.DefaultInvoiceType, pos.IsActive, pos.CreatedAt, pos.Version).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), pos)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_GetByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)

	tenantID := uuid.New()
	posID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(posColumns()).
		AddRow(posID, tenantID, 1, "Sucursal Central", true, "B", true, now, 1)

	mock.ExpectQuery("SELECT").WithArgs(posID).WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), posID)

	require.NoError(t, err)
	assert.Equal(t, posID, result.ID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, 1, result.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)
	posID := uuid.New()

	mock.ExpectQuery("SELECT").WithArgs(posID).WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), posID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)

	tenantID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(posColumns()).
		AddRow(uuid.New(), tenantID, 1, "Sucursal 1", true, "B", true, now, 1).
		AddRow(uuid.New(), tenantID, 2, "Sucursal 2", false, "C", true, now, 1)

	mock.ExpectQuery("SELECT").WithArgs(tenantID).WillReturnRows(rows)

	results, err := repo.ListByTenant(context.Background(), tenantID)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, results[0].Code)
	assert.Equal(t, 2, results[1].Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_ListByTenant_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)
	tenantID := uuid.New()

	mock.ExpectQuery("SELECT").WithArgs(tenantID).WillReturnRows(sqlmock.NewRows(posColumns()))

	results, err := repo.ListByTenant(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Empty(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_ListActiveByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)

	tenantID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows(posColumns()).
		AddRow(uuid.New(), tenantID, 1, "Sucursal Activa", true, "B", true, now, 1)

	mock.ExpectQuery("SELECT").WithArgs(tenantID).WillReturnRows(rows)

	results, err := repo.ListActiveByTenant(context.Background(), tenantID)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.True(t, results[0].IsActive)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)

	pos := entity.NewPointOfSale(uuid.New(), 1, "Sucursal Central", true, "B")
	pos.Update("Sucursal Actualizada", false, "C")

	mock.ExpectExec("UPDATE points_of_sale").
		WithArgs(pos.ID, pos.Description, pos.IsFiscalEnabled, pos.DefaultInvoiceType, pos.IsActive, pos.Version).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), pos)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPointOfSaleRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPointOfSaleRepository(db)
	pos := entity.NewPointOfSale(uuid.New(), 1, "Sucursal Central", true, "B")

	mock.ExpectExec("UPDATE points_of_sale").WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), pos)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
