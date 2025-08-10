package routes

import (
	"lystage-proj/internals/ads"

	"github.com/gin-gonic/gin"
)

func RegisterAdRoutes(rg *gin.RouterGroup) {
	adService := ads.NewAdService()
	adHandler := ads.NewHandler(adService)

	adGroup := rg.Group("/ads")
	{
		adGroup.GET("", adHandler.GetAdsHandler)
	}
}
