package controller

import (
	"log"
	"net/http"
	"tenant/src/tenant/application/command"
	"tenant/src/tenant/application/query"
	"tenant/src/tenant/application/request"
	"tenant/src/tenant/application/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantSettingsController maneja las peticiones HTTP para configuraciones estructuradas
type TenantSettingsController struct {
	getQuery      *query.GetTenantSettingsQuery
	updateCommand *command.UpdateTenantSettingsCommand
}

// NewTenantSettingsController crea una nueva instancia del controlador
func NewTenantSettingsController(
	getQuery *query.GetTenantSettingsQuery,
	updateCommand *command.UpdateTenantSettingsCommand,
) *TenantSettingsController {
	return &TenantSettingsController{
		getQuery:      getQuery,
		updateCommand: updateCommand,
	}
}

// RegisterRoutes registra las rutas del controlador
func (c *TenantSettingsController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/tenant/settings", c.GetSettings)
	router.PUT("/tenant/settings", c.UpdateSettings)
}

// GetSettings obtiene la configuración completa del tenant
// @Summary Get tenant settings
// @Description Obtiene todas las configuraciones estructuradas del tenant
// @Tags tenant-settings
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} response.TenantSettingsResponse
// @Success 404 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/settings [get]
func (c *TenantSettingsController) GetSettings(ctx *gin.Context) {
	// Obtener tenant ID del header
	tenantIDStr := ctx.GetHeader("X-Tenant-ID")
	if tenantIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "X-Tenant-ID header is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID format",
		})
		return
	}

	// Ejecutar query
	settings, err := c.getQuery.Execute(ctx.Request.Context(), tenantID)
	if err != nil {
		if err.Error() == "tenant settings not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Tenant settings not found. Please bootstrap configuration first.",
			})
			return
		}

		log.Printf("Error getting settings for tenant %s: %v", tenantID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Devolver configuración
	ctx.JSON(http.StatusOK, response.FromTenantSettings(settings))
}

// UpdateSettings actualiza la configuración del tenant
// @Summary Update tenant settings
// @Description Actualiza la configuración con optimistic locking (version control)
// @Tags tenant-settings
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body request.UpdateTenantSettingsRequest true "Settings data"
// @Success 200 {object} response.TenantSettingsResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string "Version conflict"
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/settings [put]
func (c *TenantSettingsController) UpdateSettings(ctx *gin.Context) {
	// Obtener tenant ID del header
	tenantIDStr := ctx.GetHeader("X-Tenant-ID")
	if tenantIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "X-Tenant-ID header is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID format",
		})
		return
	}

	// Parsear body
	var req request.UpdateTenantSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validar request
	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parsear cash_customer_id
	cashCustomerID, err := uuid.Parse(req.CashCustomerID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cash_customer_id format",
		})
		return
	}

	// Preparar parámetros
	params := command.UpdateTenantSettingsParams{
		TenantID:                         tenantID,
		Version:                          req.Version,
		BaseCurrency:                     req.BaseCurrency,
		AllowedCurrencies:                req.AllowedCurrencies,
		ExchangeRateSource:               req.ExchangeRateSource,
		AutoUpdateExchangeRate:           req.AutoUpdateExchangeRate,
		FiscalMode:                       req.FiscalMode,
		InvoiceGeneration:                req.InvoiceGeneration,
		AllowSaleIfAfipFails:             req.AllowSaleIfAfipFails,
		AutoRetryFailedInvoices:          req.AutoRetryFailedInvoices,
		EmailInvoiceAfterSuccess:         req.EmailInvoiceAfterSuccess,
		DefaultInvoiceType:               req.DefaultInvoiceType,
		TaxRegime:                        req.TaxRegime,
		StockPolicy:                      req.StockPolicy,
		AllowNegativeStock:               req.AllowNegativeStock,
		RequireStockValidationBeforeSale: req.RequireStockValidationBeforeSale,
		CreditEnabled:                    req.CreditEnabled,
		DefaultCreditDays:                req.DefaultCreditDays,
		MaxCreditLimit:                   req.MaxCreditLimit,
		AllowSaleOverCreditLimit:         req.AllowSaleOverCreditLimit,
		CashCustomerID:                   cashCustomerID,
	}

	// Ejecutar command
	updatedSettings, err := c.updateCommand.Execute(ctx.Request.Context(), params)
	if err != nil {
		// Version conflict
		if err.Error() == "version conflict: settings were modified by another transaction" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Version conflict",
				"message": "The settings were modified by another process. Please refresh and try again.",
			})
			return
		}

		log.Printf("Error updating settings for tenant %s: %v", tenantID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Devolver configuración actualizada
	ctx.JSON(http.StatusOK, response.FromTenantSettings(updatedSettings))
}
