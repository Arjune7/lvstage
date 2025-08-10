package routes

import (
	"lystage-proj/internals/analytics"

	"github.com/gin-gonic/gin"
)

func RegisterAnalyticsRoutes(rg *gin.RouterGroup) {
	service := analytics.NewService()
	handler := analytics.NewHandler(service)

	analyticsGroup := rg.Group("/ads")
	{
		analyticsGroup.GET("/analytics", handler.GetAdAnalytics)
	}
}
