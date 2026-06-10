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

// mockSettingsRepo for TenantSettingsController tests.
type mockSettingsRepo struct{ mock.Mock }

func (m *mockSettingsRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entity.TenantSettings, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TenantSettings), args.Error(1)
}
func (m *mockSettingsRepo) Save(ctx context.Context, settings *entity.TenantSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}
func (m *mockSettingsRepo) Exists(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.Error(1)
}

// mockEventPub reusable for controllers that need an event publisher.
type mockEventPub struct{ mock.Mock }

func (m *mockEventPub) Publish(ctx context.Context, aggID, aggType, evType string, payload []byte, by string) error {
	args := m.Called(ctx, aggID, aggType, evType, payload, by)
	return args.Error(0)
}

func setupSettingsController(repo *mockSettingsRepo, pub *mockEventPub) (*TenantSettingsController, *gin.Engine) {
	getQ := query.NewGetTenantSettingsQuery(repo)
	updateCmd := command.NewUpdateTenantSettingsCommand(repo, pub)
	ctrl := NewTenantSettingsController(getQ, updateCmd)

	r := gin.New()
	v1 := r.Group("/api/v1")
	ctrl.RegisterRoutes(v1)
	return ctrl, r
}

func validSettingsBody(cashCustomerID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"version":              1,
		"base_currency":        "ARS",
		"allowed_currencies":   []string{"ARS"},
		"exchange_rate_source": "MANUAL",
		"fiscal_mode":          "DISABLED",
		"invoice_generation":   "MANUAL",
		"default_invoice_type": "B",
		"tax_regime":           "MONOTRIBUTO",
		"stock_policy":         "IGNORE",
		"default_credit_days":  30,
		"max_credit_limit":     100,
		"cash_customer_id":     cashCustomerID.String(),
	}
}

func TestGetSettings_Found(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	tenantID := uuid.New()
	settings := entity.NewTenantSettings(tenantID, uuid.New())
	repo.On("GetByTenantID", mock.Anything, tenantID).Return(settings, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/settings", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, tenantID.String(), body["tenant_id"])
}

func TestGetSettings_NotFound(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	tenantID := uuid.New()
	repo.On("GetByTenantID", mock.Anything, tenantID).Return(nil, errors.New("tenant settings not found"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/settings", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSettings_MissingTenantHeader(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/settings", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSettings_DBError(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	tenantID := uuid.New()
	repo.On("GetByTenantID", mock.Anything, tenantID).Return(nil, errors.New("connection lost"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenant/settings", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateSettings_Success(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	tenantID := uuid.New()
	cashID := uuid.New()
	existing := entity.NewTenantSettings(tenantID, cashID)
	existing.Version = 1

	repo.On("GetByTenantID", mock.Anything, tenantID).Return(existing, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.TenantSettings")).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	bodyData := validSettingsBody(cashID)
	bodyData["version"] = 1
	bodyBytes, _ := json.Marshal(bodyData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tenant/settings", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateSettings_MissingTenantHeader(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	bodyBytes, _ := json.Marshal(validSettingsBody(uuid.New()))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tenant/settings", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSettings_InvalidBody(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tenant/settings", bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSettings_ValidationError(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	bodyData := validSettingsBody(uuid.New())
	bodyData["version"] = 0 // invalid
	bodyBytes, _ := json.Marshal(bodyData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tenant/settings", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSettings_VersionConflict(t *testing.T) {
	repo := new(mockSettingsRepo)
	pub := new(mockEventPub)
	_, r := setupSettingsController(repo, pub)

	tenantID := uuid.New()
	cashID := uuid.New()
	existing := entity.NewTenantSettings(tenantID, cashID)
	existing.Version = 5

	repo.On("GetByTenantID", mock.Anything, tenantID).Return(existing, nil)

	bodyData := validSettingsBody(cashID)
	bodyData["version"] = 1 // old version
	bodyBytes, _ := json.Marshal(bodyData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tenant/settings", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
