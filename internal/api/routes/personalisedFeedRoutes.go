package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Poll Routes
func PersonalisedFeedRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	personalisedFeedGroup := routerGroup.Group("personalised")

	personalisedFeedGroup.POST("/recompute", handler.RecomputePersonalisedFeed)
	personalisedFeedGroup.GET("/personalised", handler.FetchPersonalisedFeed)
	personalisedFeedGroup.POST("/reorder", handler.ReorderPersonalisedFeed)
}
