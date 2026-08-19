package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose Widget Routes
func WidgetRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	widgetGroup := routerGroup.Group("widget")

	widgetGroup.POST("/", handler.CreateWidget)
	widgetGroup.GET("/", handler.FetchWidget)
	widgetGroup.PUT("/:widget_id", handler.EditWidget)
	widgetGroup.DELETE("/:widget_id", handler.DeleteWidget)
	widgetGroup.POST("/lm", handler.CreateLMWidget)
	widgetGroup.PATCH("/lm/:widget_id", handler.EditLMWidget)
}
