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
	userGroup.DELETE("/:user_id", handler.DeleteUserData)
}
