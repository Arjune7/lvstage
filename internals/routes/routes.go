// internals/api/router.go
package api

import (
	"lystage-proj/internals/config"
	"lystage-proj/internals/observability"
	v1 "lystage-proj/internals/routes/v1"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(config.Logger()) // structured logs
	router.Use(observability.MetricsMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	apiGroup := router.Group("/api/v1")

	// Ads routes
	v1.RegisterAdRoutes(apiGroup)

	// Clicks routes
	v1.RegisterClickRoutes(apiGroup)

	// Analytics routes
	v1.RegisterAnalyticsRoutes(apiGroup)
	// Prometheus metrics endpoint (outside /api/v1)

	return router
}
