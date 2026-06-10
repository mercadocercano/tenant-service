package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"tenant/src/tenant/application/command"
	"tenant/src/tenant/application/query"
	"tenant/src/tenant/domain/entity"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockPOSRepo for PointOfSaleController tests.
type mockPOSRepo struct{ mock.Mock }

func (m *mockPOSRepo) Create(ctx context.Context, pos *entity.PointOfSale) error {
	args := m.Called(ctx, pos)
	return args.Error(0)
}
func (m *mockPOSRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.PointOfSale, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PointOfSale), args.Error(1)
}
func (m *mockPOSRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PointOfSale), args.Error(1)
}
func (m *mockPOSRepo) ListActiveByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.PointOfSale, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PointOfSale), args.Error(1)
}
func (m *mockPOSRepo) Update(ctx context.Context, pos *entity.PointOfSale) error {
	args := m.Called(ctx, pos)
	return args.Error(0)
}

func setupPOSController(repo *mockPOSRepo, pub *mockEventPub) (*PointOfSaleController, *gin.Engine) {
	createCmd := command.NewCreatePointOfSaleCommand(repo, pub)
	listQ := query.NewListPointsOfSaleQuery(repo)
	ctrl := NewPointOfSaleController(createCmd, listQ)

	r := gin.New()
	v1 := r.Group("/api/v1")
	ctrl.RegisterRoutes(v1)
	return ctrl, r
}

func TestCreatePointOfSale_Success(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	tenantID := uuid.New()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.PointOfSale")).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"code": 1, "description": "Sucursal Central",
		"is_fiscal_enabled": true, "default_invoice_type": "B",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/points-of-sale", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "Sucursal Central", resp["description"])
}

func TestCreatePointOfSale_MissingTenantHeader(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	body, _ := json.Marshal(map[string]interface{}{"code": 1, "description": "X", "default_invoice_type": "B"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/points-of-sale", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePointOfSale_InvalidBody(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/points-of-sale", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePointOfSale_ZeroCode_FailsValidation(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	body, _ := json.Marshal(map[string]interface{}{"code": 0, "description": "X", "default_invoice_type": "B"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/points-of-sale", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePointOfSale_RepositoryError(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	tenantID := uuid.New()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.PointOfSale")).Return(errors.New("db error"))

	body, _ := json.Marshal(map[string]interface{}{
		"code": 1, "description": "Sucursal", "default_invoice_type": "B",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/points-of-sale", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListPointsOfSale_All(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	tenantID := uuid.New()
	p1 := entity.NewPointOfSale(tenantID, 1, "Sucursal A", true, "B")
	p2 := entity.NewPointOfSale(tenantID, 2, "Sucursal B", false, "C")
	repo.On("ListByTenant", mock.Anything, tenantID).Return([]*entity.PointOfSale{p1, p2}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/points-of-sale", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, float64(2), body["total_count"])
}

func TestListPointsOfSale_OnlyActive(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	tenantID := uuid.New()
	p1 := entity.NewPointOfSale(tenantID, 1, "Sucursal A", true, "B")
	repo.On("ListActiveByTenant", mock.Anything, tenantID).Return([]*entity.PointOfSale{p1}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/points-of-sale?only_active=true", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, float64(1), body["total_count"])
}

func TestListPointsOfSale_MissingTenantHeader(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/points-of-sale", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListPointsOfSale_RepositoryError(t *testing.T) {
	repo := new(mockPOSRepo)
	pub := new(mockEventPub)
	_, r := setupPOSController(repo, pub)

	tenantID := uuid.New()
	repo.On("ListByTenant", mock.Anything, tenantID).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/points-of-sale", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
