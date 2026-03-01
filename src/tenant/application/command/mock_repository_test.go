package command

import (
	"context"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockTenantConfigRepository es un mock del repositorio
type MockTenantConfigRepository struct {
	mock.Mock
}

func (m *MockTenantConfigRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*entity.TenantConfig, bool, error) {
	args := m.Called(ctx, tenantID, key)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*entity.TenantConfig), args.Bool(1), args.Error(2)
}

func (m *MockTenantConfigRepository) Save(ctx context.Context, config *entity.TenantConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockTenantConfigRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	args := m.Called(ctx, tenantID, key)
	return args.Error(0)
}

func (m *MockTenantConfigRepository) GetAllByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantConfig, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TenantConfig), args.Error(1)
}
