package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"tenant/src/tenant/domain/port"
	tenantlog "tenant/src/tenant/infrastructure/logging"

	"github.com/stretchr/testify/assert"
)

// ADR-001: cada evento produce UNA línea JSON canónica con envelope ts/level/service/event.
func parseLine(t *testing.T, b []byte) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(b), []byte("\n"))
	assert.Len(t, lines, 1, "debe ser exactamente una línea por evento")
	var m map[string]any
	assert.NoError(t, json.Unmarshal(lines[0], &m))
	return m
}

func TestTenantLogger_SettingsBootstrapped_EnvelopeAndInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	logger.Log(port.TenantEvent{
		Event:    "tenant.settings_bootstrapped",
		TenantID: "t-123",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "tenant.settings_bootstrapped", line["event"])
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "tenant-test", line["service"])
	assert.NotEmpty(t, line["ts"], "ts (RFC3339 UTC) siempre presente")
	assert.Equal(t, "t-123", line["tenant_id"])
}

func TestTenantLogger_SettingsUpdated_InfoLevel_WithVersion(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	logger.Log(port.TenantEvent{
		Event:    "tenant.settings_updated",
		TenantID: "t-123",
		Version:  3,
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, float64(3), line["version"])
	// tenant_id es campo obligatorio de multi-tenancy
	assert.Equal(t, "t-123", line["tenant_id"])
}

func TestTenantLogger_VersionConflict_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	logger.Log(port.TenantEvent{
		Event:    "tenant.settings_version_conflict",
		TenantID: "t-123",
		Version:  2,
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "tenant.settings_version_conflict", line["event"])
}

func TestTenantLogger_SettingsUpdateFailed_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	logger.Log(port.TenantEvent{
		Event:    "tenant.settings_update_failed",
		TenantID: "t-123",
		Reason:   "db connection timeout",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "error", line["level"])
	assert.Equal(t, "db connection timeout", line["reason"])
}

func TestTenantLogger_PosCreated_InfoLevel_WithPosID(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	logger.Log(port.TenantEvent{
		Event:    "tenant.pos_created",
		TenantID: "t-123",
		PosID:    "pos-abc",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "pos-abc", line["pos_id"])
}

func TestTenantLogger_ConfigBootstrapped_OmitsEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	logger.Log(port.TenantEvent{
		Event:          "tenant.config_bootstrapped",
		TenantID:       "t-123",
		ConfigsCreated: 1,
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, float64(1), line["configs_created"])
	// omitempty: campos vacíos no aparecen
	_, hasPosID := line["pos_id"]
	assert.False(t, hasPosID, "pos_id vacío debe omitirse")
	_, hasUserID := line["user_id"]
	assert.False(t, hasUserID, "user_id vacío debe omitirse")
}

func TestTenantLogger_ConfigsCreatedZero_OmittedFromLine(t *testing.T) {
	var buf bytes.Buffer
	logger := tenantlog.NewTenantLoggerWithWriter("tenant-test", &buf)

	// configs_created=0 (el default "ya existía") no debe aparecer
	logger.Log(port.TenantEvent{
		Event:    "tenant.config_already_exists",
		TenantID: "t-123",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "warn", line["level"])
	_, hasCC := line["configs_created"]
	assert.False(t, hasCC, "configs_created=0 debe omitirse")
}
