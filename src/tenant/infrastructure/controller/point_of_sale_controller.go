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

// PointOfSaleController maneja las peticiones HTTP para puntos de venta
type PointOfSaleController struct {
	createCommand *command.CreatePointOfSaleCommand
	listQuery     *query.ListPointsOfSaleQuery
}

// NewPointOfSaleController crea una nueva instancia del controlador
func NewPointOfSaleController(
	createCommand *command.CreatePointOfSaleCommand,
	listQuery *query.ListPointsOfSaleQuery,
) *PointOfSaleController {
	return &PointOfSaleController{
		createCommand: createCommand,
		listQuery:     listQuery,
	}
}

// RegisterRoutes registra las rutas del controlador
func (c *PointOfSaleController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/tenant/points-of-sale", c.CreatePointOfSale)
	router.GET("/tenant/points-of-sale", c.ListPointsOfSale)
}

// CreatePointOfSale crea un nuevo punto de venta
// @Summary Create point of sale
// @Description Crea un nuevo punto de venta para el tenant
// @Tags points-of-sale
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body request.CreatePointOfSaleRequest true "Point of sale data"
// @Success 201 {object} response.PointOfSaleResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/points-of-sale [post]
func (c *PointOfSaleController) CreatePointOfSale(ctx *gin.Context) {
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
	var req request.CreatePointOfSaleRequest
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
	pos, err := c.createCommand.Execute(
		ctx.Request.Context(),
		tenantID,
		req.Code,
		req.Description,
		req.IsFiscalEnabled,
		req.DefaultInvoiceType,
	)

	if err != nil {
		log.Printf("Error creating point of sale for tenant %s: %v", tenantID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Devolver punto de venta creado
	ctx.JSON(http.StatusCreated, response.FromPointOfSale(pos))
}

// ListPointsOfSale lista los puntos de venta del tenant
// @Summary List points of sale
// @Description Lista todos los puntos de venta del tenant
// @Tags points-of-sale
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param only_active query boolean false "Filter only active points of sale"
// @Success 200 {array} response.PointOfSaleResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenant/points-of-sale [get]
func (c *PointOfSaleController) ListPointsOfSale(ctx *gin.Context) {
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

	// Obtener parámetro only_active
	onlyActive := ctx.Query("only_active") == "true"

	// Ejecutar query
	pointsOfSale, err := c.listQuery.Execute(ctx.Request.Context(), tenantID, onlyActive)
	if err != nil {
		log.Printf("Error listing points of sale for tenant %s: %v", tenantID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Devolver lista
	ctx.JSON(http.StatusOK, gin.H{
		"points_of_sale": response.FromPointOfSaleList(pointsOfSale),
		"total_count":    len(pointsOfSale),
	})
}
