package port

// TenantEvent es el payload canónico para eventos de dominio de tenant (ADR-001).
// Campos flat, named. Los nombres comunes (tenant_id, user_id) son idénticos al resto
// de la flota para que el LogQL cross-service funcione. Todos opcionales salvo Event.
//
// IMPORTANTE (L4): NO incluir en este struct datos de negocio sensibles del tenant
// (configs completas, secretos, credenciales fiscales). Solo IDs y campos de estado.
type TenantEvent struct {
	Event    string // <domain>.<action>_<result>, p.ej. "tenant.settings_updated"
	TenantID string
	UserID   string
	// PosID es el ID del punto de venta (solo en eventos de POS).
	PosID   string
	// ConfigKey es la clave de configuración afectada (solo en eventos de config).
	ConfigKey string
	// ConfigsCreated es la cantidad de configs creadas en un bootstrap (solo en bootstrap_config).
	ConfigsCreated int
	// Version es la versión de settings tras una actualización (solo en eventos de settings).
	Version int
	// Reason contiene el mensaje de error en eventos _failed.
	Reason string
}

// TenantEventLogger es el puerto para emitir eventos canónicos de tenant.
// El código de aplicación depende de esta interfaz; el adapter (JSON a stdout,
// Loki push, etc.) la implementa. Nunca al revés.
type TenantEventLogger interface {
	Log(e TenantEvent)
}
