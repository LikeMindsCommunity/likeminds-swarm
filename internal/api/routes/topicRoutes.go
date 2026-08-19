package routes

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Exposed Method to expose Topic Routes
func TopicRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	topicGroup := routerGroup.Group("topic")

	topicGroup.POST("/", handler.CreateTopics)
	topicGroup.GET("/", handler.FetchTopics)
	topicGroup.PUT("/:topic_id", handler.EditTopic)
	topicGroup.DELETE("/", handler.DeleteTopics)
}
