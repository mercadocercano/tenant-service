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

func init() {
	gin.SetMode(gin.TestMode)
}

// mockConfigRepo duplicates the mock from command package for controller tests.
type mockConfigRepo struct{ mock.Mock }

func (m *mockConfigRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*entity.TenantConfig, bool, error) {
	args := m.Called(ctx, tenantID, key)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*entity.TenantConfig), args.Bool(1), args.Error(2)
}
func (m *mockConfigRepo) Save(ctx context.Context, cfg *entity.TenantConfig) error {
	args := m.Called(ctx, cfg)
	return args.Error(0)
}
func (m *mockConfigRepo) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	args := m.Called(ctx, tenantID, key)
	return args.Error(0)
}
func (m *mockConfigRepo) GetAllByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantConfig, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TenantConfig), args.Error(1)
}

func setupConfigController(repo *mockConfigRepo) (*TenantConfigController, *gin.Engine) {
	getQ := query.NewGetTenantConfigQuery(repo)
	setCmd := command.NewSetTenantConfigCommand(repo)
	bootCmd := command.NewBootstrapTenantConfigCommand(repo)
	ctrl := NewTenantConfigController(getQ, setCmd, bootCmd)

	r := gin.New()
	v1 := r.Group("/api/v1")
	ctrl.RegisterRoutes(v1)
	return ctrl, r
}

func TestGetConfig_Found(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	tenantID := uuid.New()
	cfg := entity.NewTenantConfig(tenantID, "catalog.stock_policy", "REQUIRE_STOCK")
	repo.On("GetByKey", mock.Anything, tenantID, "catalog.stock_policy").Return(cfg, true, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/config/catalog.stock_policy", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "catalog.stock_policy", body["key"])
	assert.Equal(t, "REQUIRE_STOCK", body["value"])
}

func TestGetConfig_NotFound(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	tenantID := uuid.New()
	repo.On("GetByKey", mock.Anything, tenantID, "unknown.key").Return(nil, false, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/config/unknown.key", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "unknown.key", body["key"])
	assert.Nil(t, body["value"])
}

func TestGetConfig_MissingTenantHeader(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/config/some.key", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConfig_InvalidTenantID(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/config/some.key", nil)
	req.Header.Set("X-Tenant-ID", "not-a-uuid")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConfig_RepositoryError(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	tenantID := uuid.New()
	repo.On("GetByKey", mock.Anything, tenantID, "some.key").Return(nil, false, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/config/some.key", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetConfig_Success_NewKey(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	tenantID := uuid.New()
	repo.On("GetByKey", mock.Anything, tenantID, "catalog.stock_policy").Return(nil, false, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.TenantConfig")).Return(nil)

	body, _ := json.Marshal(map[string]string{"key": "catalog.stock_policy", "value": "REQUIRE_STOCK"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "catalog.stock_policy", resp["key"])
	assert.Equal(t, "REQUIRE_STOCK", resp["value"])
}

func TestSetConfig_MissingTenantHeader(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	body, _ := json.Marshal(map[string]string{"key": "k", "value": "v"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetConfig_InvalidBody(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/config", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetConfig_EmptyKey_FailsValidation(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	body, _ := json.Marshal(map[string]string{"key": "", "value": "v"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBootstrapConfig_Success(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	tenantID := uuid.New()
	repo.On("GetByKey", mock.Anything, tenantID, "catalog.stock_policy").Return(nil, false, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.TenantConfig")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/bootstrap", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, float64(1), body["configs_created"])
}

func TestBootstrapConfig_Idempotent(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	tenantID := uuid.New()
	cfg := entity.NewTenantConfig(tenantID, "catalog.stock_policy", "REQUIRE_STOCK")
	repo.On("GetByKey", mock.Anything, tenantID, "catalog.stock_policy").Return(cfg, true, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/bootstrap", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, float64(0), body["configs_created"])
}

func TestBootstrapConfig_MissingTenantHeader(t *testing.T) {
	repo := new(mockConfigRepo)
	_, r := setupConfigController(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tenant/bootstrap", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
