package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose Base Routes
func BaseRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	baseGroup := routerGroup.Group("")

	baseGroup.GET("/comment/:comment_id", handler.FetchCommentById)
	baseGroup.GET("/comment", handler.FetchComments)
	baseGroup.DELETE("/cache", handler.DeleteCache)
}
