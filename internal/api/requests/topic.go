package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

// Request Structure for Create Topic
type CreateTopicRequest struct {
	Names []string `json:"names"`
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

// Response Structure for Fetch Topic Response
type FetchTopicResponse struct {
	ID            primitive.ObjectID `json:"_id"`
	Name          string             `json:"name"`
	IsEnabled     bool               `json:"is_enabled"`
	NumberOfPosts int                `json:"number_of_posts"`
}

// Request Structure for Fetch Topic
type FetchTopicRequest struct {
	Search     string `form:"search"`
	SearchType string `form:"search_type"`
	IsEnabled  string `form:"is_enabled"`
	MinPosts   int    `json:"min_posts"`
}

// Request Structure for Delete Topics
type DeleteTopicsRequest struct {
	TopicIds []string `json:"topic_ids"`
}
