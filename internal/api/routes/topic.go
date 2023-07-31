package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
)

// Exposed Method to expose Topic Routes
func TopicRouter(routerGroup *gin.RouterGroup, handler *handlers.FeedHandlers) {
	topicGroup := routerGroup.Group("topic")

	topicGroup.POST("/", handler.CreateTopic)
	topicGroup.GET("/", handler.FetchTopics)
	topicGroup.PUT("/:topic_id", handler.EditTopic)
}
