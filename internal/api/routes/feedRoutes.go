package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose Feed Routes
func FeedRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	feedGroup := routerGroup.Group("feed")

	feedGroup.GET("/universal", handler.FetchUniversalFeed)
	feedGroup.GET("/explore", handler.FetchExploreFeed)
	feedGroup.GET("/group", handler.FetchGroupFeed)
	feedGroup.GET("/connection", handler.FetchConnectionFeed)
}
