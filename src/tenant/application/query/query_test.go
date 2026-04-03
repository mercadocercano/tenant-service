package query

import (
	"context"
	"errors"
	"testing"
	"tenant/src/tenant/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

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

// --- GetTenantConfigQuery Tests ---

func TestGetTenantConfigQuery_Execute_Found(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	key := "catalog.stock_policy"
	mockRepo := new(MockTenantConfigRepository)

	expectedConfig := entity.NewTenantConfig(tenantID, key, "IGNORE_STOCK")
	mockRepo.On("GetByKey", ctx, tenantID, key).Return(expectedConfig, true, nil)

	query := NewGetTenantConfigQuery(mockRepo)

	// Act
	config, exists, err := query.Execute(ctx, tenantID, key)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, config)
	assert.Equal(t, "IGNORE_STOCK", config.Value)
	mockRepo.AssertExpectations(t)
}

func TestGetTenantConfigQuery_Execute_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	key := "nonexistent.key"
	mockRepo := new(MockTenantConfigRepository)

	mockRepo.On("GetByKey", ctx, tenantID, key).Return(nil, false, nil)

	query := NewGetTenantConfigQuery(mockRepo)

	// Act
	config, exists, err := query.Execute(ctx, tenantID, key)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, config)
}

func TestGetTenantConfigQuery_Execute_Error(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	key := "catalog.stock_policy"
	mockRepo := new(MockTenantConfigRepository)

	mockRepo.On("GetByKey", ctx, tenantID, key).Return(nil, false, errors.New("db error"))

	query := NewGetTenantConfigQuery(mockRepo)

	// Act
	config, exists, err := query.Execute(ctx, tenantID, key)

	// Assert
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Nil(t, config)
}

// --- GetTenantSettingsQuery Tests ---

func TestGetTenantSettingsQuery_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	expectedSettings := entity.NewTenantSettings(tenantID, uuid.New())
	mockRepo.On("GetByTenantID", ctx, tenantID).Return(expectedSettings, nil)

	query := NewGetTenantSettingsQuery(mockRepo)

	// Act
	settings, err := query.Execute(ctx, tenantID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, tenantID, settings.TenantID)
	mockRepo.AssertExpectations(t)
}

func TestGetTenantSettingsQuery_Execute_Error(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockTenantSettingsRepository)

	mockRepo.On("GetByTenantID", ctx, tenantID).Return(nil, errors.New("not found"))

	query := NewGetTenantSettingsQuery(mockRepo)

	// Act
	settings, err := query.Execute(ctx, tenantID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, settings)
}

// --- ListPointsOfSaleQuery Tests ---

func TestListPointsOfSaleQuery_Execute_AllPoints(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	activePOS := entity.NewPointOfSale(tenantID, 1, "Activa", true, "B")
	inactivePOS := entity.NewPointOfSale(tenantID, 2, "Inactiva", true, "B")
	inactivePOS.Deactivate()

	allPOS := []*entity.PointOfSale{activePOS, inactivePOS}
	mockRepo.On("ListByTenant", ctx, tenantID).Return(allPOS, nil)

	query := NewListPointsOfSaleQuery(mockRepo)

	// Act
	result, err := query.Execute(ctx, tenantID, false)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestListPointsOfSaleQuery_Execute_OnlyActive(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	activePOS := entity.NewPointOfSale(tenantID, 1, "Activa", true, "B")
	mockRepo.On("ListActiveByTenant", ctx, tenantID).Return([]*entity.PointOfSale{activePOS}, nil)

	query := NewListPointsOfSaleQuery(mockRepo)

	// Act
	result, err := query.Execute(ctx, tenantID, true)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.True(t, result[0].IsActive)
	mockRepo.AssertExpectations(t)
}

func TestListPointsOfSaleQuery_Execute_Empty(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	mockRepo.On("ListByTenant", ctx, tenantID).Return([]*entity.PointOfSale{}, nil)

	query := NewListPointsOfSaleQuery(mockRepo)

	// Act
	result, err := query.Execute(ctx, tenantID, false)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestListPointsOfSaleQuery_Execute_Error(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tenantID := uuid.New()
	mockRepo := new(MockPointOfSaleRepository)

	mockRepo.On("ListByTenant", ctx, tenantID).Return(nil, errors.New("db error"))

	query := NewListPointsOfSaleQuery(mockRepo)

	// Act
	result, err := query.Execute(ctx, tenantID, false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}
