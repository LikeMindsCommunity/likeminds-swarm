package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Poll Routes
func PollRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	pollGroup := routerGroup.Group("poll")

	pollGroup.PUT("/:poll_id", handler.AddPollOption)
	pollGroup.PUT("/:poll_id/vote", handler.VoteOnPoll)
	pollGroup.GET("/:poll_id/vote", handler.GetPollVotes)
}
