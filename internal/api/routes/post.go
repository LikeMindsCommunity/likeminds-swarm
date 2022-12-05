package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
)

func PostRouter(routerGroup *gin.RouterGroup, postHelper interfaces.PostHelper, likeHelper interfaces.LikeHelper,
	commentHelper interfaces.CommentHelper, saveHelper interfaces.SaveHelper, activityHelper interfaces.ActivityHelper) {
	postHandlers := handlers.NewPostHandlers(postHelper)
	commentHandlers := handlers.NewCommentHandlers(commentHelper, postHelper)
	likeHandlers := handlers.NewLikeHandlers(postHelper, likeHelper, commentHelper, activityHelper)
	saveHandlers := handlers.NewSaveHandlers(saveHelper, postHelper)

	postGroup := routerGroup.Group("post")
	postGroup.POST("/", postHandlers.CreatePost)
	postGroup.GET("/:post_id", postHandlers.FetchPost)
	postGroup.DELETE("/:post_id", postHandlers.DeletePost)
	postGroup.PUT("/:post_id/pin", postHandlers.PinPost)
	postGroup.GET("/:post_id/like", likeHandlers.FetchPostLikes)
	postGroup.PUT("/:post_id/like", likeHandlers.LikePost)
	postGroup.PUT("/:post_id/save", saveHandlers.SavePost)
	postGroup.POST("/:post_id/comment/*comment_id", commentHandlers.CommentPostOrComment)
	postGroup.DELETE("/:post_id/comment/:comment_id", commentHandlers.DeleteComment)
	postGroup.PUT("/:post_id/comment/:comment_id/like", likeHandlers.LikeComment)
}
