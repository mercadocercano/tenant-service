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

// TenantConfigController maneja las peticiones HTTP para configuraciones
type TenantConfigController struct {
	getConfigQuery   *query.GetTenantConfigQuery
	setConfigCommand *command.SetTenantConfigCommand
	bootstrapCommand *command.BootstrapTenantConfigCommand
}

// NewTenantConfigController crea una nueva instancia del controlador
func NewTenantConfigController(
	getConfigQuery *query.GetTenantConfigQuery,
	setConfigCommand *command.SetTenantConfigCommand,
	bootstrapCommand *command.BootstrapTenantConfigCommand,
) *TenantConfigController {
	return &TenantConfigController{
		getConfigQuery:   getConfigQuery,
		setConfigCommand: setConfigCommand,
		bootstrapCommand: bootstrapCommand,
	}
}

// RegisterRoutes registra las rutas del controlador
func (c *TenantConfigController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/tenant/config/:key", c.GetConfig)
	router.POST("/tenant/config", c.SetConfig)
	router.POST("/tenant/bootstrap", c.BootstrapConfig)
}

// GetConfig obtiene una configuración por clave
// @Summary Get tenant configuration by key
// @Description Obtiene el valor de una configuración específica del tenant
// @Tags tenant-config
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param key path string true "Configuration Key (e.g., catalog.stock_policy)"
// @Success 200 {object} response.SimpleConfigResponse
// @Success 404 {object} response.SimpleConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/config/{key} [get]
func (c *TenantConfigController) GetConfig(ctx *gin.Context) {
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

	// Obtener key del path
	key := ctx.Param("key")
	if key == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Configuration key is required",
		})
		return
	}

	// Ejecutar query
	config, exists, err := c.getConfigQuery.Execute(ctx.Request.Context(), tenantID, key)
	if err != nil {
		log.Printf("Error getting config for tenant %s, key %s: %v", tenantID, key, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Si no existe, devolver 404 con value null
	if !exists {
		ctx.JSON(http.StatusNotFound, response.NewSimpleResponse(key, nil))
		return
	}

	// Devolver configuración encontrada
	ctx.JSON(http.StatusOK, response.NewSimpleResponse(key, &config.Value))
}

// SetConfig establece o actualiza una configuración
// @Summary Set tenant configuration
// @Description Crea o actualiza una configuración del tenant (upsert)
// @Tags tenant-config
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body request.SetTenantConfigRequest true "Configuration data"
// @Success 200 {object} response.TenantConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/config [post]
func (c *TenantConfigController) SetConfig(ctx *gin.Context) {
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
	var req request.SetTenantConfigRequest
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

	// Ejecutar command
	config, err := c.setConfigCommand.Execute(ctx.Request.Context(), tenantID, req.Key, req.Value)
	if err != nil {
		log.Printf("Error setting config for tenant %s, key %s: %v", tenantID, req.Key, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Devolver configuración guardada
	ctx.JSON(http.StatusOK, response.FromEntity(config))
}

// BootstrapConfig inicializa la configuración default de un tenant
// @Summary Bootstrap tenant configuration
// @Description Crea configuraciones por defecto para un tenant nuevo (idempotente)
// @Tags tenant-config
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/bootstrap [post]
func (c *TenantConfigController) BootstrapConfig(ctx *gin.Context) {
	log.Printf("=== BOOTSTRAP ENDPOINT START ===")

	// Obtener tenant ID del header
	tenantIDStr := ctx.GetHeader("X-Tenant-ID")
	if tenantIDStr == "" {
		log.Printf("ERROR: X-Tenant-ID header is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "X-Tenant-ID header is required",
		})
		return
	}

	log.Printf("TenantID from header: %s", tenantIDStr)

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid tenant ID format: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID format",
		})
		return
	}

	// Ejecutar bootstrap command
	log.Printf("Executing bootstrap command for tenant: %s", tenantID)
	createdCount, err := c.bootstrapCommand.Execute(ctx.Request.Context(), tenantID)
	if err != nil {
		log.Printf("ERROR: Bootstrap command failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	log.Printf("Bootstrap completed successfully: %d configs created", createdCount)
	log.Printf("=== BOOTSTRAP ENDPOINT END ===")

	// Siempre responder 200 (idempotente)
	ctx.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Tenant configuration bootstrapped successfully",
		"tenant_id":       tenantID.String(),
		"configs_created": createdCount,
	})
}
