package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Custom Widget Routes
func CustomWidgetRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	customWidgetGroup := routerGroup.Group("widget")

	customWidgetGroup.POST("/", handler.CreateCustomWidget)
	customWidgetGroup.GET("/", handler.FetchCustomWidget)
	customWidgetGroup.PUT("/:widget_id", handler.EditCustomWidget)
}
