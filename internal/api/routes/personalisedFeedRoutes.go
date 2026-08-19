package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose Poll Routes
func PersonalisedFeedRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	personalisedFeedGroup := routerGroup.Group("personalised")

	personalisedFeedGroup.GET("/", handler.FetchPersonalisedFeed)
	personalisedFeedGroup.POST("/recompute", handler.RecomputePersonalisedFeed)
	personalisedFeedGroup.POST("/reorder", handler.ReorderPersonalisedFeed)
	personalisedFeedGroup.POST("/compute/community/default", handler.ComputeCommunityDefaultFeed)
}
