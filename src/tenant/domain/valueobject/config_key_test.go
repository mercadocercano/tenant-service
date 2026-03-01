package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigKey_Valid(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:      "valid namespaced key",
			value:     "catalog.stock_policy",
			wantError: false,
		},
		{
			name:      "valid nested key",
			value:     "integrations.mercadopago.api_key",
			wantError: false,
		},
		{
			name:      "empty key",
			value:     "",
			wantError: true,
		},
		{
			name:      "no namespace",
			value:     "stockpolicy",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := NewConfigKey(tt.value)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
				assert.Equal(t, tt.value, key.String())
			}
		})
	}
}

func TestConfigKey_Namespace(t *testing.T) {
	key, _ := NewConfigKey("catalog.stock_policy")
	assert.Equal(t, "catalog", key.Namespace())
}

func TestConfigKey_Key(t *testing.T) {
	key, _ := NewConfigKey("catalog.stock_policy")
	assert.Equal(t, "stock_policy", key.Key())
}

func TestConfigKey_NestedKey(t *testing.T) {
	key, _ := NewConfigKey("integrations.mercadopago.api_key")
	assert.Equal(t, "integrations", key.Namespace())
	assert.Equal(t, "mercadopago.api_key", key.Key())
}
