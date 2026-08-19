package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose User Routes
func UserRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	userGroup := routerGroup.Group("user")

	userGroup.GET("/:user_id/save", handler.FetchUserSavedPosts)
	userGroup.GET("/:user_id/post", handler.FetchUserCreatedPosts)
	userGroup.GET("/:user_id/post/pending", handler.FetchUserCreatedPendingPosts)
	userGroup.GET("/:user_id/comment", handler.FetchUserComments)

	userGroup.GET("/activity", handler.FetchNotificationFeed)
	userGroup.POST("/activity", handler.ExternalCreateNotificationActivity)

	userGroup.POST("/activity/:activity_id/mark_read", handler.NotificationFeedActivityMarkRead)
	userGroup.GET("/activity/unread_count", handler.NotificationFeedUnreadCount)

	userGroup.GET("/:user_id/activity", handler.FetchUserProfileActivity)

	userGroup.PATCH("/:user_id/connection", handler.UpdateConnection)

	userGroup.GET("/:user_id/meta", handler.FetchUserFeedMeta)

	userGroup.GET("/topics", handler.FetchUsersTopics)
	userGroup.PATCH("/:user_id/topics", handler.UpdateUserTopics)

	userGroup.DELETE("/", handler.DeleteUserData)
}
