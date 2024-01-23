package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Base Routes
func BaseRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	baseGroup := routerGroup.Group("")

	baseGroup.GET("/comment/:comment_id", handler.FetchCommentById)
	baseGroup.GET("/comment", handler.FetchComments)
	baseGroup.DELETE("/cache", handler.DeleteCache)
}
