package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose Post Routes
func PostRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	postGroup := routerGroup.Group("post")

	postGroup.POST("/", handler.CreatePost)
	postGroup.GET("/", handler.FetchPosts)
	postGroup.PUT("/:post_id", handler.EditPost)
	postGroup.GET("/:post_id", handler.FetchPost)
	postGroup.DELETE("/:post_id", handler.DeletePost)
	postGroup.PUT("/:post_id/pin", handler.PinPost)
	postGroup.PUT("/:post_id/hide", handler.HidePost)
	postGroup.GET("/:post_id/like", handler.FetchPostLikes)
	postGroup.PUT("/:post_id/like", handler.LikePost)
	postGroup.PUT("/:post_id/save", handler.SavePost)

	postGroup.POST("/seen", handler.MarkPostsSeen)

	postGroup.POST("/:post_id/comment", handler.CommentPost)
	postGroup.PUT("/:post_id/comment/:comment_id", handler.EditComment)
	postGroup.GET("/:post_id/comment/:comment_id", handler.FetchComment)
	postGroup.DELETE("/:post_id/comment/:comment_id", handler.DeleteComment)
	postGroup.POST("/:post_id/comment/:comment_id/comment", handler.ReplyComment)
	postGroup.GET("/:post_id/comment/:comment_id/like", handler.FetchCommentLikes)
	postGroup.PUT("/:post_id/comment/:comment_id/like", handler.LikeComment)

	postGroup.PUT("/:post_id/share/count", handler.UpdatePostShareCount)

	postGroup.GET("/search", handler.SearchPost)
	postGroup.GET("/search/user/:user_id", handler.SearchUserCreatedPost)

	postGroup.POST("/pending", handler.CreatePendingPostForReview)
	postGroup.PATCH("/pending/:pending_post_id", handler.ApproveOrRejectPendingPost)
	postGroup.PUT("/pending/:pending_post_id", handler.EditPendingPost)
	postGroup.GET("/pending/:pending_post_id", handler.FetchPendingPost)
	postGroup.DELETE("/pending/:pending_post_id", handler.DeletePendingPost)
}
