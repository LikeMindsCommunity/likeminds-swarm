package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose User Routes
func UserRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	userGroup := routerGroup.Group("user")

	userGroup.GET("/:user_id/save", handler.FetchUserSavedPosts)
	userGroup.GET("/:user_id/post", handler.FetchUserCreatedPosts)
	userGroup.GET("/:user_id/activity", handler.FetchUserActivity)
	userGroup.POST("/:user_id/activity", handler.ExternalCreateActivity)
	userGroup.POST("/:user_id/activity/:activity_id/mark_read", handler.UserActivityMarkRead)
	userGroup.GET("/:user_id/activity/unread_count", handler.UserActivityFeedUnreadCount)
	userGroup.PATCH("/:user_id/connection", handler.UpdateConnection)
	userGroup.DELETE("/", handler.DeleteUserData)
}
