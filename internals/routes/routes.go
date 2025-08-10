// internals/api/router.go
package api

import (
	"lystage-proj/internals/config"
	v1 "lystage-proj/internals/routes/v1"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(config.Logger()) // structured logs

	apiGroup := router.Group("/api/v1")

	// Ads routes
	v1.RegisterAdRoutes(apiGroup)

	// Clicks routes
	v1.RegisterClickRoutes(apiGroup)

	// Analytics routes
	v1.RegisterAnalyticsRoutes(apiGroup)

	return router
}
