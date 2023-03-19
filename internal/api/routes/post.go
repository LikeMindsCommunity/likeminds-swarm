package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Post Routes
func PostRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	postGroup := routerGroup.Group("post")

	postGroup.POST("/", handler.CreatePost)
	postGroup.GET("/:post_id", handler.FetchPost)
	postGroup.DELETE("/:post_id", handler.DeletePost)
	postGroup.PUT("/:post_id/pin", handler.PinPost)
	postGroup.GET("/:post_id/like", handler.FetchPostLikes)
	postGroup.PUT("/:post_id/like", handler.LikePost)
	postGroup.PUT("/:post_id/save", handler.SavePost)
	postGroup.GET("/:post_id/comment/:comment_id", handler.FetchComment)
	postGroup.POST("/:post_id/comment", handler.CommentPost)
	postGroup.POST("/:post_id/comment/:comment_id/comment", handler.ReplyComment)
	postGroup.DELETE("/:post_id/comment/:comment_id", handler.DeleteComment)
	postGroup.GET("/:post_id/comment/:comment_id/like", handler.FetchCommentLikes)
	postGroup.PUT("/:post_id/comment/:comment_id/like", handler.LikeComment)
	postGroup.GET("/search", handler.SearchPost)
}
