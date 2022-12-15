package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

func UserRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	postGroup := routerGroup.Group("user")

	postGroup.GET("/:user_id/save", handler.FetchUserSavedPosts)
	postGroup.GET("/:user_id/post", handler.FetchUserCreatedPosts)
	postGroup.POST("/:user_id/activity", handler.ExternalCreateActivity)
}
