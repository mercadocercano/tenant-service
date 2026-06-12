package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	localConfig "tenant/src/shared/infrastructure/config"
	tenantConfig "tenant/src/tenant/infrastructure/config"
	"tenant/src/tenant/infrastructure/event"

	"github.com/gin-gonic/gin"
	sharedConfig "github.com/hornosg/go-shared/infrastructure/config"
	"github.com/hornosg/go-shared/infrastructure/env"
	tenantmw "github.com/hornosg/go-shared/infrastructure/middleware"
	"github.com/hornosg/go-shared/infrastructure/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mercadocercano/eventbus"
)

func main() {
	// Configurar el router con Gin
	router := gin.New()

	// Agregar middlewares básicos necesarios
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(tenantmw.TenantValidation(tenantmw.TenantValidationConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		ExcludedRoutes: []string{
			"/health",
			"/metrics",
		},
	}))

	// Configurar Prometheus metrics si está habilitado
	prometheusEnabled := os.Getenv("PROMETHEUS_ENABLED")
	log.Printf("PROMETHEUS_ENABLED value: '%s'", prometheusEnabled)

	if prometheusEnabled == "true" {
		log.Println("Registering /metrics endpoint for Tenant service")
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
		log.Println("/metrics endpoint registered successfully for Tenant service")
	} else {
		log.Println("Prometheus metrics disabled for Tenant service")
	}

	// Configurar middlewares compartidos (Gzip via go-shared + CORS local)
	gzipCfg := sharedConfig.SharedConfig{
		EnableGzip:          true,
		AlwaysTryDecompress: true,
		GzipExcludedPaths:   []string{"/health", "/metrics"},
	}
	sharedConfig.SetupSharedMiddleware(router, gzipCfg)

	corsCfg := localConfig.DefaultCORSConfig()
	localConfig.SetupCORSMiddleware(router, corsCfg)

	// Obtener configuración de la base de datos de variables de entorno
	dbHost := env.Get("DB_HOST", "localhost")
	dbPort := env.Get("DB_PORT", "5432")
	dbUser := env.Get("DB_USER", "postgres")
	dbPassword := env.Get("DB_PASSWORD", "postgres")
	dbName := env.Get("DB_NAME", "tenant_db")

	log.Printf("Intentando conectar a postgres://%s:***@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	// Conectar a la base de datos
	db, err := postgres.Connect(postgres.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()
	log.Println("Conexión a la base de datos establecida con éxito")

	postgres.StartPoolMonitor(context.Background(), db, postgres.MonitorOptions{Service: "tenant-service", DBName: dbName})

	// Conectar a la base de datos del eventbus
	eventBusHost := env.Get("EVENTBUS_DB_HOST", "localhost")
	eventBusPort := env.Get("EVENTBUS_DB_PORT", "5432")
	eventBusUser := env.Get("EVENTBUS_DB_USER", "postgres")
	eventBusPassword := env.Get("EVENTBUS_DB_PASSWORD", "postgres")
	eventBusName := env.Get("EVENTBUS_DB_NAME", "eventbus")

	log.Printf("Conectando a EventBus en postgres://%s:***@%s:%s/%s", eventBusUser, eventBusHost, eventBusPort, eventBusName)

	eventBusDB, err := postgres.Connect(postgres.Config{
		Host:     eventBusHost,
		Port:     eventBusPort,
		User:     eventBusUser,
		Password: eventBusPassword,
		DBName:   eventBusName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos del eventbus: %v", err)
	}
	defer eventBusDB.Close()
	log.Println("Conexión al eventbus establecida con éxito")

	postgres.StartPoolMonitor(context.Background(), eventBusDB, postgres.MonitorOptions{Service: "tenant-service", DBName: eventBusName})

	// Configurar eventbus publisher
	logger := eventbus.NewLogger(eventbus.LevelInfo)
	eventStore := eventbus.NewSQLEventStore(eventBusDB, logger)
	publishUseCase := eventbus.NewPublishEventUseCase(eventStore, logger)
	eventPublisher := event.NewEventPublisherAdapter(publishUseCase)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "tenant-service",
			"version": "1.0.0",
		})
	})

	// API v1 grupo de rutas
	v1 := router.Group("/api/v1")

	// Configurar módulo tenant (con eventbus)
	setupTenantModule(v1, db, eventPublisher)

	// Obtener puerto del entorno
	port := env.Get("PORT", "8120")

	// Iniciar el servidor
	log.Printf("Servidor iniciando en http://localhost:%s", port)
	router.Run(":" + port)
}

// setupTenantModule configura el módulo Tenant
func setupTenantModule(router *gin.RouterGroup, db *sql.DB, eventPublisher *event.EventPublisherAdapter) {
	log.Println("Configurando módulo Tenant...")

	// Crear configuración completa del módulo Tenant con eventbus
	tenantCfg := tenantConfig.NewExtendedTenantModuleConfig(db, eventPublisher)

	// Registrar rutas existentes (key-value)
	tenantCfg.ConfigController.RegisterRoutes(router)

	// Registrar nuevas rutas (estructuradas)
	tenantCfg.SettingsController.RegisterRoutes(router)
	tenantCfg.PointOfSaleController.RegisterRoutes(router)

	log.Println("Módulo Tenant configurado exitosamente")
	log.Println("Rutas Tenant disponibles:")
	log.Println("  [Key-Value Config]")
	log.Println("  GET    /api/v1/tenant/config/:key")
	log.Println("  POST   /api/v1/tenant/config")
	log.Println("  POST   /api/v1/tenant/bootstrap")
	log.Println("  [Structured Settings]")
	log.Println("  GET    /api/v1/tenant/settings")
	log.Println("  PUT    /api/v1/tenant/settings")
	log.Println("  [Points of Sale]")
	log.Println("  POST   /api/v1/tenant/points-of-sale")
	log.Println("  GET    /api/v1/tenant/points-of-sale")
}
