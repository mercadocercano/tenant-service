package config

import (
	"github.com/gin-gonic/gin"
)

// SharedConfig contiene la configuración para módulos compartidos
type SharedConfig struct {
	EnableCORS bool
}

// DefaultSharedConfig devuelve una configuración por defecto
func DefaultSharedConfig() SharedConfig {
	return SharedConfig{
		EnableCORS: true,
	}
}

// SetupSharedMiddleware configura los middlewares compartidos
func SetupSharedMiddleware(router *gin.Engine, config SharedConfig) {
	// CORS básico para desarrollo
	if config.EnableCORS {
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})
	}
}
