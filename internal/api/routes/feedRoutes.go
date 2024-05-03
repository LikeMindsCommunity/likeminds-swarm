package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Feed Routes
func FeedRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	feedGroup := routerGroup.Group("feed")

	feedGroup.GET("/universal", handler.FetchUniversalFeed)
	feedGroup.GET("/explore", handler.FetchExploreFeed)
	feedGroup.GET("/group", handler.FetchGroupFeed)
	feedGroup.GET("/connection", handler.FetchConnectionFeed)
}
