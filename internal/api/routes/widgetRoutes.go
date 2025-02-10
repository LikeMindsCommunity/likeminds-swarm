package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
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
