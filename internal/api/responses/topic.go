package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

// Response Structure for Topic
type TopicResponse struct {
	ID           primitive.ObjectID `json:"_id"`
	Name         string             `json:"name"`
	IsEnabled    bool               `json:"is_enabled"`
	Priority     float32            `json:"priority"`
	IsSearchable bool               `json:"is_searchable"`
	ParentId     primitive.ObjectID `json:"parent_id,omitempty"`
	ParentName   string             `json:"parent_name,omitempty"`
	Level        int                `json:"level"`
	WidgetId     primitive.ObjectID `json:"widget_id,omitempty"`
}

// Response Structure for fetched Indexed Topics Response
type FetchTopicsResponse struct {
	*TopicResponse
	NumberOfPosts   int `json:"number_of_posts"`
	TotalChildCount int `json:"total_child_count"`
}
