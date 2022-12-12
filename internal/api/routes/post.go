package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
)

func PostRouter(routerGroup *gin.RouterGroup, postHelper interfaces.PostHelper, likeHelper interfaces.LikeHelper,
	commentHelper interfaces.CommentHelper, saveHelper interfaces.SaveHelper, activityHelper interfaces.ActivityHelper) {
	postHandlers := handlers.NewPostHandlers(postHelper, likeHelper, commentHelper, activityHelper, saveHelper)
	commentHandlers := handlers.NewCommentHandlers(commentHelper, likeHelper, postHelper, activityHelper)
	likeHandlers := handlers.NewLikeHandlers(postHelper, likeHelper, commentHelper, activityHelper)
	saveHandlers := handlers.NewSaveHandlers(saveHelper, likeHelper, commentHelper, postHelper)

	postGroup := routerGroup.Group("post")
	postGroup.POST("/", postHandlers.CreatePost)
	postGroup.GET("/:post_id", postHandlers.FetchPost)
	postGroup.DELETE("/:post_id", postHandlers.DeletePost)
	postGroup.PUT("/:post_id/pin", postHandlers.PinPost)
	postGroup.GET("/:post_id/like", likeHandlers.FetchPostLikes)
	postGroup.PUT("/:post_id/like", likeHandlers.LikePost)
	postGroup.PUT("/:post_id/save", saveHandlers.SavePost)
	postGroup.GET("/:post_id/comment/:comment_id", commentHandlers.FetchComment)
	postGroup.POST("/:post_id/comment", commentHandlers.CommentPost)
	postGroup.POST("/:post_id/comment/:comment_id/comment", commentHandlers.ReplyComment)
	postGroup.DELETE("/:post_id/comment/:comment_id", commentHandlers.DeleteComment)
	postGroup.GET("/:post_id/comment/:comment_id/like", likeHandlers.FetchCommentLikes)
	postGroup.PUT("/:post_id/comment/:comment_id/like", likeHandlers.LikeComment)
}
