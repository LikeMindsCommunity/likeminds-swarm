package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
)

func UserRouter(routerGroup *gin.RouterGroup, activityHelper interfaces.ActivityHelper, likeHelper interfaces.LikeHelper,
	commentHelper interfaces.CommentHelper, postHelper interfaces.PostHelper, saveHelper interfaces.SaveHelper) {
	postHandlers := handlers.NewPostHandlers(postHelper, likeHelper, commentHelper, activityHelper, saveHelper)
	activityHandlers := handlers.NewActivityHandlers(activityHelper)
	saveHandlers := handlers.NewSaveHandlers(saveHelper, likeHelper, commentHelper, postHelper)

	postGroup := routerGroup.Group("user")
	postGroup.GET("/:user_id/save", saveHandlers.FetchUserSavedPosts)
	postGroup.GET("/:user_id/post", postHandlers.FetchUserCreatedPosts)
	postGroup.POST("/:user_id/activity", activityHandlers.ExternalCreateActivity)
}
