//go:build integration

// Test de aislamiento cross-tenant adversarial (PLAT-E29 T8).
//
// tenant_config, tenant_settings y points_of_sale son el centro de configuración operativa y
// fiscal del tenant (AFIP, monedas, límites de crédito). Este test es el criterio REAL de
// fail-closed de la épica: no que la migración 005 corra, sino que un tenant jamás vea ni
// escriba filas de otro bajo RLS.
//
// Levanta un Postgres efímero (testcontainers), aplica TODAS las migraciones de esquema reales
// del servicio (001, 003, 004, 005 — el hueco 002 es benigno, y el subdir migrations/roles/
// queda afuera vía e.IsDir()) y conecta bajo un rol sin BYPASSRLS que replica tenant_app
// (creado en lab-postgres por PLAT-E29 T3, sin sus grants del eventbus: los repos de esta
// épica nunca tocan esa conexión). Un superuser SIEMPRE bypasea RLS aunque la tabla tenga
// FORCE ROW LEVEL SECURITY, así que probar contra el usuario `postgres` del contenedor daría
// falsos verdes (gotcha probado empíricamente en E24 T3/T6). TODAS las aserciones corren bajo
// tenant_app_test (NOBYPASSRLS) → las policies tenant_isolation de la 005 se ejercen de verdad.
//
// Cobertura (T8 "Hecho cuando", patrón E25–E28 adaptado a las 3 tablas de tenant-service):
//
//	(a) Save/Create de A bajo RLSContext{A} pasa el WITH CHECK en las 3 tablas.
//	(b) los datos de A NO son visibles bajo RLSContext{B}: GetByKey, GetByTenantID,
//	    ListByTenant devuelven vacío/not-found, y query cruda por PK/clave también.
//	(c) INSERT con tenant_id=A bajo sesión B → rechazado por WITH CHECK (las 3 tablas).
//	(d) UPDATE cruda cross-tenant de B sobre filas de A → 0 filas, fila intacta.
//	(e) fail-closed sin contexto (M1): SELECT/INSERT/UPDATE directos a las 3 tablas sin
//	    app.tenant_id fijado erran o devuelven 0 filas — lecturas Y escrituras.
//	(f) upsert de tenant_config (ON CONFLICT (tenant_id, config_key)) bajo B con la misma
//	    config_key de A: NO pisa la fila de A (crea la propia de B).
//	(g) optimistic locking de tenant_settings: el flujo update-then-insert de Save funciona
//	    bajo RLS para el tenant propio (version bump) y no puede tocar la fila ajena.
//
// El contenedor se levanta UNA sola vez para todo el binario vía TestMain (no por TestXxx).
//
// Para correrlo:
//
//	GOWORK=off go test -tags=integration ./test/tenant/infrastructure/persistence/rls/... -count=1 -v
package rls_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/hornosg/go-shared/infrastructure/postgres"

	"tenant/src/tenant/domain/entity"
	"tenant/src/tenant/infrastructure/persistence"
)

// containerDB es el nombre de la base del contenedor efímero. El rol de app necesita CONNECT
// sobre ESTA base (tenant_app.sql hardcodea `GRANT CONNECT ON DATABASE tenant_db`, el nombre
// real de lab-postgres, que no existe acá; por eso ese script NO se aplica y el rol se
// replica abajo).
const containerDB = "tenant_test"

// appRoleName/appRolePassword replican en el Postgres efímero el rol tenant_app creado en
// lab-postgres por PLAT-E29 T3 (grants sobre tenant_db únicamente — el eventbus queda fuera:
// ninguno de los 3 repositorios de esta épica escribe en esa conexión).
const (
	appRoleName     = "tenant_app_test"
	appRolePassword = "tenant_app_test"
)

// appDB es la única conexión que usan los TestXxx de este archivo — bajo el rol sin BYPASSRLS,
// la que de verdad queda sujeta a las policies tenant_isolation de la 005.
var appDB *sql.DB

// TestMain levanta el Postgres efímero, aplica las migraciones de esquema reales y crea el rol
// UNA sola vez para todo el binario — se comparte entre todas las TestXxx en vez de pagar el
// arranque del contenedor por cada una.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(containerDB),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("error starting postgres container: %v", err)
	}

	superConnStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("error getting connection string: %v", err)
	}

	superDB, err := sql.Open("postgres", superConnStr)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	if err := superDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	if err := applyMigrations(superDB); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}
	if err := createRoles(superDB); err != nil {
		log.Fatalf("error creating roles: %v", err)
	}

	appConnStr, err := withCredentials(superConnStr, appRoleName, appRolePassword)
	if err != nil {
		log.Fatalf("error building app connection string: %v", err)
	}
	appDB, err = sql.Open("postgres", appConnStr)
	if err != nil {
		log.Fatalf("error opening app-role database: %v", err)
	}
	if err := appDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database as %s: %v", appRoleName, err)
	}

	code := m.Run()

	_ = appDB.Close()
	_ = superDB.Close()
	_ = container.Terminate(ctx)

	os.Exit(code)
}

// applyMigrations corre en orden TODAS las migraciones .up.sql de ESQUEMA de tenant-service
// (001 tenant_config + 003 tenant_settings + 004 points_of_sale + 005 RLS — el hueco 002 es
// benigno, golang-migrate versiona por número de archivo, no por secuencia contigua) — el mismo
// esquema que corre en lab-postgres/tenant_db tras T2. El subdirectorio migrations/roles/
// (tenant_app.sql) queda afuera: os.ReadDir lo devuelve como dir (e.IsDir() → skip) y además
// hardcodea `tenant_db` + grants del eventbus, que no existen/aplican en el contenedor efímero.
// El rol de app se replica acá vía createRoles() con sus propios grants.
func applyMigrations(db *sql.DB) error {
	dir := "../../../../../migrations"
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range dirEntries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".sql" || strings.HasSuffix(name, ".down.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return errors.New("no se encontraron migraciones .up.sql en " + dir)
	}

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return errors.New("migración " + f + ": " + err.Error())
		}
	}
	return nil
}

// createRoles crea (con el superuser) el rol de aplicación sin DDL ni BYPASSRLS con los mismos
// grants efectivos que tenant_app.sql le da a tenant_app en lab-postgres sobre tenant_db (menos
// el SELECT sobre schema_migrations, que este apply directo no crea, y menos los grants del
// eventbus, que esta épica nunca ejerce desde los repositorios) — réplica del rol real, no un
// stand-in: SELECT/INSERT/UPDATE sin DELETE en tenant_config/tenant_settings, SELECT/INSERT sin
// UPDATE/DELETE en points_of_sale (D2: Update fue podado en T4).
func createRoles(superDB *sql.DB) error {
	stmts := []string{
		`CREATE ROLE ` + appRoleName + ` WITH LOGIN PASSWORD '` + appRolePassword + `' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS`,
		`GRANT CONNECT ON DATABASE ` + containerDB + ` TO ` + appRoleName,
		`GRANT USAGE ON SCHEMA public TO ` + appRoleName,
		`REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC`,
		`GRANT SELECT, INSERT, UPDATE ON tenant_config TO ` + appRoleName,
		`GRANT SELECT, INSERT, UPDATE ON tenant_settings TO ` + appRoleName,
		`GRANT SELECT, INSERT ON points_of_sale TO ` + appRoleName,
	}
	for _, stmt := range stmts {
		if _, err := superDB.Exec(stmt); err != nil {
			return errors.New("stmt " + stmt + ": " + err.Error())
		}
	}
	return nil
}

// withCredentials reconstruye la connection string del contenedor reemplazando usuario y
// contraseña — evita depender de las APIs internas de testcontainers para host/puerto.
func withCredentials(connStr, user, password string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(user, password)
	return u.String(), nil
}

// countRaw cuenta filas bajo un RLSContext dado — usado para las aserciones de invisibilidad
// cross-tenant (nunca vía el WHERE tenant_id=... del repository, para no confundir aislamiento
// por RLS con un filtro de la query).
func countRaw(t *testing.T, rc postgres.RLSContext, query string, args ...interface{}) int {
	t.Helper()
	var count int
	err := postgres.WithRLSInTransaction(context.Background(), appDB, rc, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(&count)
	})
	require.NoError(t, err)
	return count
}

// TestTenantConfig_CrossTenantIsolation cubre (a)-(d) y (f) sobre tenant_config: la cadena
// fail-closed bajo el rol NOBYPASSRLS — lectura, escritura y el upsert ON CONFLICT
// (tenant_id, config_key), que debe seguir siendo per-tenant bajo RLS.
func TestTenantConfig_CrossTenantIsolation(t *testing.T) {
	repo := persistence.NewPostgresTenantConfigRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	configA := entity.NewTenantConfig(tenantA, "catalog.stock_policy", "REQUIRE_STOCK")
	require.NoError(t, repo.Save(ctx, configA), "el Save del propio tenant debe pasar el WITH CHECK")

	t.Run("(a) A lee su config por el repository", func(t *testing.T) {
		got, found, err := repo.GetByKey(ctx, tenantA, "catalog.stock_policy")
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, "REQUIRE_STOCK", got.Value)
	})

	t.Run("(b) B NO ve la config de A ni por el repository ni por consulta cruda", func(t *testing.T) {
		_, found, err := repo.GetByKey(ctx, tenantB, "catalog.stock_policy")
		require.NoError(t, err)
		require.False(t, found, "la config de A es visible desde B")

		count := countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM tenant_config WHERE id = $1`, configA.ID)
		require.Equal(t, 0, count, "la fila de A es visible por consulta cruda desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK", func(t *testing.T) {
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, e := tx.ExecContext(ctx,
					`INSERT INTO tenant_config (id, tenant_id, config_key, config_value, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
					uuid.New(), tenantA, "intruso.key", "x")
				return e
			})
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		var affected int64
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				res, e := tx.ExecContext(ctx,
					`UPDATE tenant_config SET config_value = 'hackeado' WHERE id = $1`, configA.ID)
				if e != nil {
					return e
				}
				affected, e = res.RowsAffected()
				return e
			})
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		got, found, err := repo.GetByKey(ctx, tenantA, "catalog.stock_policy")
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, "REQUIRE_STOCK", got.Value, "el config_value de A no debe haber cambiado por el UPDATE de B")
	})

	t.Run("(f) upsert de B con la misma config_key de A no pisa la fila de A", func(t *testing.T) {
		configB := entity.NewTenantConfig(tenantB, "catalog.stock_policy", "IGNORE")
		require.NoError(t, repo.Save(ctx, configB), "el Save (upsert) del propio tenant debe pasar el WITH CHECK")

		gotB, found, err := repo.GetByKey(ctx, tenantB, "catalog.stock_policy")
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, "IGNORE", gotB.Value)

		gotA, found, err := repo.GetByKey(ctx, tenantA, "catalog.stock_policy")
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, "REQUIRE_STOCK", gotA.Value, "el upsert de B con la misma config_key no debe afectar la fila de A")
	})
}

// TestTenantSettings_CrossTenantIsolation cubre (a)-(d) y (g) sobre tenant_settings: lectura,
// escritura, y el flujo update-then-insert con optimistic locking de Save() bajo RLS.
func TestTenantSettings_CrossTenantIsolation(t *testing.T) {
	repo := persistence.NewPostgresTenantSettingsRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	settingsA := entity.NewTenantSettings(tenantA, uuid.New())
	require.NoError(t, repo.Save(ctx, settingsA), "el Save (insert inicial) del propio tenant debe pasar el WITH CHECK")

	t.Run("(a) A lee su settings por el repository", func(t *testing.T) {
		got, err := repo.GetByTenantID(ctx, tenantA)
		require.NoError(t, err)
		require.Equal(t, tenantA, got.TenantID)
		require.Equal(t, "ARS", got.BaseCurrency)
	})

	t.Run("(b) B NO ve el settings de A ni por el repository ni por Exists ni por consulta cruda", func(t *testing.T) {
		_, err := repo.GetByTenantID(ctx, tenantB)
		require.Error(t, err, "GetByTenantID de B no debe encontrar el settings de A")

		exists, err := repo.Exists(ctx, tenantB)
		require.NoError(t, err)
		require.False(t, exists, "Exists bajo B no debe ver el settings de A")

		count := countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM tenant_settings WHERE tenant_id = $1`, tenantA)
		require.Equal(t, 0, count, "el settings de A es visible por consulta cruda desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK", func(t *testing.T) {
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, e := tx.ExecContext(ctx,
					`INSERT INTO tenant_settings (
						tenant_id, base_currency, allowed_currencies, exchange_rate_source,
						auto_update_exchange_rate, fiscal_mode, invoice_generation,
						allow_sale_if_afip_fails, auto_retry_failed_invoices,
						email_invoice_after_success, default_invoice_type, tax_regime,
						stock_policy, allow_negative_stock, require_stock_validation_before_sale,
						credit_enabled, default_credit_days, max_credit_limit,
						allow_sale_over_credit_limit, cash_customer_id, version, updated_at
					) VALUES (
						$1, 'ARS', '["ARS"]', 'MANUAL', false, 'DISABLED', 'MANUAL',
						true, false, false, 'B', 'MONOTRIBUTO',
						'IGNORE', true, false,
						false, 30, 0,
						false, $2, 1, NOW()
					)`,
					tenantA, uuid.New())
				return e
			})
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		var affected int64
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				res, e := tx.ExecContext(ctx,
					`UPDATE tenant_settings SET base_currency = 'USD' WHERE tenant_id = $1`, tenantA)
				if e != nil {
					return e
				}
				affected, e = res.RowsAffected()
				return e
			})
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		got, err := repo.GetByTenantID(ctx, tenantA)
		require.NoError(t, err)
		require.Equal(t, "ARS", got.BaseCurrency, "el base_currency de A no debe haber cambiado por el UPDATE de B")
	})

	t.Run("(g) optimistic locking: A puede actualizar su propio settings (version bump) bajo RLS", func(t *testing.T) {
		got, err := repo.GetByTenantID(ctx, tenantA)
		require.NoError(t, err)
		require.Equal(t, 1, got.Version)

		got.IncrementVersion()
		got.BaseCurrency = "USD"
		got.AllowedCurrencies = []string{"ARS", "USD"}
		require.NoError(t, repo.Save(ctx, got), "el Save (update-then-insert) con version correcta debe pasar bajo RLS")

		updated, err := repo.GetByTenantID(ctx, tenantA)
		require.NoError(t, err)
		require.Equal(t, "USD", updated.BaseCurrency)
		require.Equal(t, 2, updated.Version)
	})
}

// TestPointOfSale_CrossTenantIsolation cubre (a)-(d) sobre points_of_sale.
func TestPointOfSale_CrossTenantIsolation(t *testing.T) {
	repo := persistence.NewPostgresPointOfSaleRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	posA := entity.NewPointOfSale(tenantA, 1, "Sucursal A", true, "B")
	require.NoError(t, repo.Create(ctx, posA), "el Create del propio tenant debe pasar el WITH CHECK")

	t.Run("(a) A ve su punto de venta por ListByTenant/ListActiveByTenant", func(t *testing.T) {
		list, err := repo.ListByTenant(ctx, tenantA)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, posA.ID, list[0].ID)

		active, err := repo.ListActiveByTenant(ctx, tenantA)
		require.NoError(t, err)
		require.Len(t, active, 1)
	})

	t.Run("(b) B NO ve el punto de venta de A ni por el repository ni por consulta cruda", func(t *testing.T) {
		list, err := repo.ListByTenant(ctx, tenantB)
		require.NoError(t, err)
		require.Empty(t, list, "el punto de venta de A es visible desde B")

		count := countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM points_of_sale WHERE id = $1`, posA.ID)
		require.Equal(t, 0, count, "la fila de A es visible por consulta cruda desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK", func(t *testing.T) {
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, e := tx.ExecContext(ctx,
					`INSERT INTO points_of_sale (id, tenant_id, code, description, is_fiscal_enabled, default_invoice_type, is_active, created_at, version)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), 1)`,
					uuid.New(), tenantA, 99, "intruso", true, "B", true)
				return e
			})
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A falla (D2: sin grant UPDATE en points_of_sale)", func(t *testing.T) {
		// D2 podó PointOfSaleRepository.Update (T4): el rol tenant_app NO tiene GRANT UPDATE
		// sobre points_of_sale (T3), así que la defensa acá es least-privilege — más estricta
		// que la policy RLS (ni siquiera el propio tenant puede hacer UPDATE crudo).
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, e := tx.ExecContext(ctx,
					`UPDATE points_of_sale SET is_active = false WHERE id = $1`, posA.ID)
				return e
			})
		require.Error(t, err, "esperaba permission denied: tenant_app no tiene GRANT UPDATE en points_of_sale (D2)")

		active, err := repo.ListActiveByTenant(ctx, tenantA)
		require.NoError(t, err)
		require.Len(t, active, 1, "el punto de venta de A debe seguir activo pese al intento de UPDATE de B")
	})
}

// TestFailClosed_WithoutTenantContext cubre (e) / M1: lecturas Y escrituras directas al pool
// (fuera de WithRLSInTransaction → sin SET LOCAL app.tenant_id) erran o devuelven 0 filas en
// LAS 3 TABLAS — nunca todas. Corre reutilizando el pool appDB cuyas conexiones físicas YA
// tuvieron SET LOCAL fijado (y reseteado al commitear) — el escenario real de un pool
// reutilizado entre requests de distintos tenants.
func TestFailClosed_WithoutTenantContext(t *testing.T) {
	ctx := context.Background()

	tenant := uuid.New()
	repoConfig := persistence.NewPostgresTenantConfigRepository(appDB)
	repoSettings := persistence.NewPostgresTenantSettingsRepository(appDB)
	repoPOS := persistence.NewPostgresPointOfSaleRepository(appDB)

	// Garantizar al menos una fila en cada tabla para que "0 visibles" sea significativo.
	require.NoError(t, repoConfig.Save(ctx, entity.NewTenantConfig(tenant, "existe.key", "v")))
	require.NoError(t, repoSettings.Save(ctx, entity.NewTenantSettings(tenant, uuid.New())))
	require.NoError(t, repoPOS.Create(ctx, entity.NewPointOfSale(tenant, 1, "existe", true, "B")))

	t.Run("(e) lecturas sin contexto: 0 filas o error, nunca todas", func(t *testing.T) {
		for _, table := range []string{"tenant_config", "tenant_settings", "points_of_sale"} {
			var count int
			err := appDB.QueryRowContext(ctx, `SELECT count(*) FROM `+table).Scan(&count)
			if err != nil {
				continue // fail-closed vía error (current_setting sin valor): comportamiento esperado
			}
			require.Equal(t, 0, count, table+" devolvió filas sin contexto de tenant — RLS no está aislando")
		}
	})

	t.Run("(e) escrituras sin contexto fallan", func(t *testing.T) {
		_, errInsConfig := appDB.ExecContext(ctx,
			`INSERT INTO tenant_config (id, tenant_id, config_key, config_value, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())`, uuid.New(), tenant, "nueva.key", "v")
		require.Error(t, errInsConfig, "INSERT en tenant_config sin app.tenant_id debe fallar")

		_, errUpdConfig := appDB.ExecContext(ctx,
			`UPDATE tenant_config SET config_value = 'x' WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdConfig, "UPDATE en tenant_config sin app.tenant_id debe fallar")

		_, errInsSettings := appDB.ExecContext(ctx,
			`INSERT INTO tenant_settings (
				tenant_id, base_currency, allowed_currencies, exchange_rate_source,
				auto_update_exchange_rate, fiscal_mode, invoice_generation,
				allow_sale_if_afip_fails, auto_retry_failed_invoices,
				email_invoice_after_success, default_invoice_type, tax_regime,
				stock_policy, allow_negative_stock, require_stock_validation_before_sale,
				credit_enabled, default_credit_days, max_credit_limit,
				allow_sale_over_credit_limit, cash_customer_id, version, updated_at
			) VALUES (
				$1, 'ARS', '["ARS"]', 'MANUAL', false, 'DISABLED', 'MANUAL',
				true, false, false, 'B', 'MONOTRIBUTO',
				'IGNORE', true, false,
				false, 30, 0,
				false, $2, 1, NOW()
			)`, uuid.New(), uuid.New())
		require.Error(t, errInsSettings, "INSERT en tenant_settings sin app.tenant_id debe fallar")

		_, errUpdSettings := appDB.ExecContext(ctx,
			`UPDATE tenant_settings SET base_currency = 'USD' WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdSettings, "UPDATE en tenant_settings sin app.tenant_id debe fallar")

		_, errInsPOS := appDB.ExecContext(ctx,
			`INSERT INTO points_of_sale (id, tenant_id, code, description, is_fiscal_enabled, default_invoice_type, is_active, created_at, version)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), 1)`,
			uuid.New(), tenant, 2, "nuevo", true, "B", true)
		require.Error(t, errInsPOS, "INSERT en points_of_sale sin app.tenant_id debe fallar")

		_, errUpdPOS := appDB.ExecContext(ctx,
			`UPDATE points_of_sale SET is_active = false WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdPOS, "UPDATE en points_of_sale sin app.tenant_id debe fallar")
	})
}
