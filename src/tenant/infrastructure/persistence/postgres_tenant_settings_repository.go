package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/domain/repository"

	"github.com/google/uuid"
	"github.com/hornosg/go-shared/infrastructure/postgres"
	"github.com/lib/pq"
)

// PostgresTenantSettingsRepository implementa el repositorio usando PostgreSQL.
//
// RLS (E29, RULE-09/RULE-10): `tenant_settings` tiene ROW LEVEL SECURITY forzado con la
// policy `tenant_isolation` (migración 005). Cada operación corre dentro de
// postgres.WithRLSInTransaction, que fija `app.tenant_id` con SET LOCAL — sin él, cualquier
// query erra (fail-closed) bajo el rol NOBYPASSRLS `tenant_app`. El filtro manual
// `WHERE tenant_id = $` se mantiene como defensa en profundidad. Save() envuelve el flujo
// completo update-then-insert (+ optimistic locking) en UNA sola transacción RLS: el intento
// de INSERT/UPDATE y la resolución de conflicto comparten el mismo `app.tenant_id`.
type PostgresTenantSettingsRepository struct {
	db *sql.DB
}

// NewPostgresTenantSettingsRepository crea una nueva instancia del repositorio
func NewPostgresTenantSettingsRepository(db *sql.DB) repository.TenantSettingsRepository {
	return &PostgresTenantSettingsRepository{db: db}
}

// GetByTenantID obtiene la configuración de un tenant. Tenant: tenantID param.
func (r *PostgresTenantSettingsRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entity.TenantSettings, error) {
	query := `
		SELECT
			tenant_id,
			base_currency,
			allowed_currencies,
			exchange_rate_source,
			auto_update_exchange_rate,
			fiscal_mode,
			invoice_generation,
			allow_sale_if_afip_fails,
			auto_retry_failed_invoices,
			email_invoice_after_success,
			default_invoice_type,
			tax_regime,
			stock_policy,
			allow_negative_stock,
			require_stock_validation_before_sale,
			credit_enabled,
			default_credit_days,
			max_credit_limit,
			allow_sale_over_credit_limit,
			cash_customer_id,
			version,
			updated_at
		FROM tenant_settings
		WHERE tenant_id = $1
	`

	rc := postgres.RLSContext{TenantID: tenantID.String()}
	var settings entity.TenantSettings
	var allowedCurrenciesJSON []byte

	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, tenantID).Scan(
			&settings.TenantID,
			&settings.BaseCurrency,
			&allowedCurrenciesJSON,
			&settings.ExchangeRateSource,
			&settings.AutoUpdateExchangeRate,
			&settings.FiscalMode,
			&settings.InvoiceGeneration,
			&settings.AllowSaleIfAfipFails,
			&settings.AutoRetryFailedInvoices,
			&settings.EmailInvoiceAfterSuccess,
			&settings.DefaultInvoiceType,
			&settings.TaxRegime,
			&settings.StockPolicy,
			&settings.AllowNegativeStock,
			&settings.RequireStockValidationBeforeSale,
			&settings.CreditEnabled,
			&settings.DefaultCreditDays,
			&settings.MaxCreditLimit,
			&settings.AllowSaleOverCreditLimit,
			&settings.CashCustomerID,
			&settings.Version,
			&settings.UpdatedAt,
		)
	})

	if err == sql.ErrNoRows {
		return nil, errors.New("tenant settings not found")
	}

	if err != nil {
		return nil, err
	}

	// Deserializar allowed_currencies
	if err := json.Unmarshal(allowedCurrenciesJSON, &settings.AllowedCurrencies); err != nil {
		return nil, err
	}

	return &settings, nil
}

// Save guarda o actualiza la configuración con optimistic locking. Tenant: settings.TenantID
// (path del WITH CHECK). Todo el flujo update-then-insert corre en una única transacción RLS
// — el UPDATE, la resolución de conflicto y el INSERT comparten el mismo `app.tenant_id`.
func (r *PostgresTenantSettingsRepository) Save(ctx context.Context, settings *entity.TenantSettings) error {
	// Serializar allowed_currencies a JSON
	allowedCurrenciesJSON, err := json.Marshal(settings.AllowedCurrencies)
	if err != nil {
		return err
	}

	rc := postgres.RLSContext{TenantID: settings.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		// Intentar UPDATE primero (si existe)
		updateQuery := `
			UPDATE tenant_settings
			SET
				base_currency = $2,
				allowed_currencies = $3,
				exchange_rate_source = $4,
				auto_update_exchange_rate = $5,
				fiscal_mode = $6,
				invoice_generation = $7,
				allow_sale_if_afip_fails = $8,
				auto_retry_failed_invoices = $9,
				email_invoice_after_success = $10,
				default_invoice_type = $11,
				tax_regime = $12,
				stock_policy = $13,
				allow_negative_stock = $14,
				require_stock_validation_before_sale = $15,
				credit_enabled = $16,
				default_credit_days = $17,
				max_credit_limit = $18,
				allow_sale_over_credit_limit = $19,
				cash_customer_id = $20,
				version = $21,
				updated_at = $22
			WHERE tenant_id = $1 AND version = $23
		`

		previousVersion := settings.Version - 1 // La versión que debe estar en DB

		result, err := tx.ExecContext(ctx, updateQuery,
			settings.TenantID,
			settings.BaseCurrency,
			allowedCurrenciesJSON,
			settings.ExchangeRateSource,
			settings.AutoUpdateExchangeRate,
			settings.FiscalMode,
			settings.InvoiceGeneration,
			settings.AllowSaleIfAfipFails,
			settings.AutoRetryFailedInvoices,
			settings.EmailInvoiceAfterSuccess,
			settings.DefaultInvoiceType,
			settings.TaxRegime,
			settings.StockPolicy,
			settings.AllowNegativeStock,
			settings.RequireStockValidationBeforeSale,
			settings.CreditEnabled,
			settings.DefaultCreditDays,
			settings.MaxCreditLimit,
			settings.AllowSaleOverCreditLimit,
			settings.CashCustomerID,
			settings.Version,
			settings.UpdatedAt,
			previousVersion,
		)

		if err != nil {
			// Si es violación de clave primaria, intentar INSERT
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return r.insertTx(ctx, tx, settings, allowedCurrenciesJSON)
			}
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		// Si no afectó filas, puede ser:
		// 1. No existe (hacer INSERT)
		// 2. Version conflict
		if rowsAffected == 0 {
			// Verificar si existe (misma tx: no re-entrar por WithRLSInTransaction)
			var exists bool
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS(SELECT 1 FROM tenant_settings WHERE tenant_id = $1)`,
				settings.TenantID,
			).Scan(&exists); err != nil {
				return err
			}

			if exists {
				// Existe pero version no coincide
				return errors.New("version conflict: settings were modified by another transaction")
			}

			// No existe, hacer INSERT
			return r.insertTx(ctx, tx, settings, allowedCurrenciesJSON)
		}

		return nil
	})
}

// insertTx realiza la inserción inicial dentro de la transacción RLS abierta por Save.
func (r *PostgresTenantSettingsRepository) insertTx(ctx context.Context, tx *sql.Tx, settings *entity.TenantSettings, allowedCurrenciesJSON []byte) error {
	insertQuery := `
		INSERT INTO tenant_settings (
			tenant_id,
			base_currency,
			allowed_currencies,
			exchange_rate_source,
			auto_update_exchange_rate,
			fiscal_mode,
			invoice_generation,
			allow_sale_if_afip_fails,
			auto_retry_failed_invoices,
			email_invoice_after_success,
			default_invoice_type,
			tax_regime,
			stock_policy,
			allow_negative_stock,
			require_stock_validation_before_sale,
			credit_enabled,
			default_credit_days,
			max_credit_limit,
			allow_sale_over_credit_limit,
			cash_customer_id,
			version,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22
		)
	`

	_, err := tx.ExecContext(ctx, insertQuery,
		settings.TenantID,
		settings.BaseCurrency,
		allowedCurrenciesJSON,
		settings.ExchangeRateSource,
		settings.AutoUpdateExchangeRate,
		settings.FiscalMode,
		settings.InvoiceGeneration,
		settings.AllowSaleIfAfipFails,
		settings.AutoRetryFailedInvoices,
		settings.EmailInvoiceAfterSuccess,
		settings.DefaultInvoiceType,
		settings.TaxRegime,
		settings.StockPolicy,
		settings.AllowNegativeStock,
		settings.RequireStockValidationBeforeSale,
		settings.CreditEnabled,
		settings.DefaultCreditDays,
		settings.MaxCreditLimit,
		settings.AllowSaleOverCreditLimit,
		settings.CashCustomerID,
		settings.Version,
		settings.UpdatedAt,
	)

	return err
}

// Exists verifica si existe configuración para un tenant. Tenant: tenantID param.
func (r *PostgresTenantSettingsRepository) Exists(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenant_settings WHERE tenant_id = $1)`

	rc := postgres.RLSContext{TenantID: tenantID.String()}
	var exists bool
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, tenantID).Scan(&exists)
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}
