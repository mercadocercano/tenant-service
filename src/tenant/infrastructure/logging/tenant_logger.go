package logging

import (
	"io"

	"tenant/src/tenant/domain/port"

	sharedlog "github.com/hornosg/go-shared/infrastructure/logging"
)

// TenantLogger implementa port.TenantEventLogger emitiendo una línea JSON canónica
// (ADR-001) por evento, delegando el envelope (ts/level/service/event + campos flat
// omitempty) en go-shared CanonicalLogger (>= v0.8.0). El mapeo struct→fields y las
// reglas de nivel por evento viven acá; el formato canónico es compartido por la flota.
//
// IMPORTANTE (L4): este adapter NUNCA registra datos de negocio sensibles del tenant
// (configs completas, secretos fiscales, credenciales). Solo IDs y campos de estado.
type TenantLogger struct {
	canonical *sharedlog.CanonicalLogger
}

// NewTenantLogger crea el adapter escribiendo a stdout. El service se fija acá, nunca por-call.
func NewTenantLogger(service string) *TenantLogger {
	return &TenantLogger{canonical: sharedlog.NewCanonicalLogger(service)}
}

// NewTenantLoggerWithWriter permite inyectar un io.Writer (tests).
func NewTenantLoggerWithWriter(service string, w io.Writer) *TenantLogger {
	return &TenantLogger{canonical: sharedlog.NewCanonicalLoggerWithWriter(service, w)}
}

// levelFor aplica las reglas de nivel del ADR-001 por tipo de evento.
func levelFor(event string) string {
	switch event {
	// Flujo normal — info
	case "tenant.settings_bootstrapped",
		"tenant.settings_updated",
		"tenant.config_bootstrapped",
		"tenant.config_set",
		"tenant.pos_created":
		return "info"
	// Conflicto de versión / ya existe — warn (anomalía recuperable)
	case "tenant.settings_version_conflict",
		"tenant.settings_already_exists",
		"tenant.config_already_exists":
		return "warn"
	// Errores de persistencia / validación — error
	case "tenant.settings_bootstrap_failed",
		"tenant.settings_update_failed",
		"tenant.config_bootstrap_failed",
		"tenant.config_set_failed",
		"tenant.pos_create_failed":
		return "error"
	default:
		return "info"
	}
}

// Log emite una línea JSON canónica para el evento dado.
func (l *TenantLogger) Log(e port.TenantEvent) {
	fields := map[string]any{
		"tenant_id":  e.TenantID,
		"user_id":    e.UserID,
		"pos_id":     e.PosID,
		"config_key": e.ConfigKey,
		"reason":     e.Reason,
	}
	if e.ConfigsCreated > 0 {
		fields["configs_created"] = e.ConfigsCreated
	}
	if e.Version > 0 {
		fields["version"] = e.Version
	}
	l.canonical.Emit(levelFor(e.Event), e.Event, fields)
}
