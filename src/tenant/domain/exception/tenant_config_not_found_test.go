package exception

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenantConfigNotFound_Error(t *testing.T) {
	err := NewTenantConfigNotFound("tenant-abc", "catalog.stock_policy")

	assert.Equal(t, "tenant config not found for tenant tenant-abc with key catalog.stock_policy", err.Error())
}

func TestTenantConfigNotFound_Fields(t *testing.T) {
	err := NewTenantConfigNotFound("t-123", "fiscal.mode")

	assert.Equal(t, "t-123", err.TenantID)
	assert.Equal(t, "fiscal.mode", err.Key)
}

func TestTenantConfigNotFound_ImplementsError(t *testing.T) {
	var err error = NewTenantConfigNotFound("t-1", "k")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "tenant config not found")
}
