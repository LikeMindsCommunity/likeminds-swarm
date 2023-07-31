package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

// Request Structure for Create Topic
type CreateTopicRequest struct {
	Name string `json:"name" binding:"required"`
}

// Request Structure for Edit Topic
type EditTopicRequest struct {
	Name      string `json:"name"`
	IsEnabled bool   `json:"is_enabled"`
}

// Response Structure for Topic
type TopicResponse struct {
	ID        primitive.ObjectID `json:"_id"`
	Name      string             `json:"name"`
	IsEnabled bool               `json:"is_enabled"`
}
