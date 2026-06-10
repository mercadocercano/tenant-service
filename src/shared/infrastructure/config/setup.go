package config

import (
	"github.com/gin-gonic/gin"
)

// CORSConfig contiene la configuración CORS para el tenant-service
type CORSConfig struct {
	EnableCORS bool
}

// DefaultCORSConfig devuelve la configuración CORS por defecto
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		EnableCORS: true,
	}
}

// SetupCORSMiddleware configura el middleware CORS básico para desarrollo
func SetupCORSMiddleware(router *gin.Engine, config CORSConfig) {
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
