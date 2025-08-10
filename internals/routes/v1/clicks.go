package routes

import (
	"lystage-proj/internals/clicks"

	"github.com/gin-gonic/gin"
)

func RegisterClickRoutes(r *gin.RouterGroup) {
	service := clicks.NewService()
	handler := clicks.NewHandler(service)

	clickGroup := r.Group("/ads")
	{
		clickGroup.POST("/click", handler.HandleClick)
	}
}
