package command

import (
	"context"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockTenantSettingsRepository es un mock del repositorio de settings
type MockTenantSettingsRepository struct {
	mock.Mock
}

func (m *MockTenantSettingsRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entity.TenantSettings, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TenantSettings), args.Error(1)
}

func (m *MockTenantSettingsRepository) Save(ctx context.Context, settings *entity.TenantSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockTenantSettingsRepository) Exists(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.Error(1)
}

// MockPointOfSaleRepository es un mock del repositorio de puntos de venta
type MockPointOfSaleRepository struct {
	mock.Mock
}

func (m *MockPointOfSaleRepository) Create(ctx context.Context, pos *entity.PointOfSale) error {
	args := m.Called(ctx, pos)
	return args.Error(0)
}

func (m *MockPointOfSaleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PointOfSale, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PointOfSale), args.Error(1)
}

func (m *MockPointOfSaleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PointOfSale), args.Error(1)
}

func (m *MockPointOfSaleRepository) ListActiveByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PointOfSale), args.Error(1)
}

func (m *MockPointOfSaleRepository) Update(ctx context.Context, pos *entity.PointOfSale) error {
	args := m.Called(ctx, pos)
	return args.Error(0)
}

// MockEventPublisher es un mock del publicador de eventos
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, aggregateID, aggregateType, eventType string, payload []byte, publishedBy string) error {
	args := m.Called(ctx, aggregateID, aggregateType, eventType, payload, publishedBy)
	return args.Error(0)
}
