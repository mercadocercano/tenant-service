package response

import (
	"tenant/src/tenant/domain/entity"
	"time"

	"github.com/google/uuid"
)

// TenantConfigResponse representa la respuesta de una configuración
type TenantConfigResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromEntity crea una respuesta desde una entidad
func FromEntity(config *entity.TenantConfig) *TenantConfigResponse {
	return &TenantConfigResponse{
		ID:        config.ID,
		TenantID:  config.TenantID,
		Key:       config.Key,
		Value:     config.Value,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}
}

// SimpleConfigResponse representa una respuesta simplificada (solo key y value)
type SimpleConfigResponse struct {
	Key   string  `json:"key"`
	Value *string `json:"value"` // Nullable para indicar que no existe
}

// NewSimpleResponse crea una respuesta simplificada
func NewSimpleResponse(key string, value *string) *SimpleConfigResponse {
	return &SimpleConfigResponse{
		Key:   key,
		Value: value,
	}
}
