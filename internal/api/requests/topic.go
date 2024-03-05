package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateTopicRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Priority     float32                `json:"priority"`
	IsSearchable *bool                  `json:"is_searchable"`
	ParentId     string                 `json:"parent_id"`
	IsEnabled    *bool                  `json:"is_enabled"`
	Metadata     map[string]interface{} `json:"metadata"`

	ParentName      string               `json:"-"` // For Internal use
	AllParentIds    []primitive.ObjectID `json:"-"`
	Level           int                  `json:"-"`
	WidgetId        primitive.ObjectID   `json:"-"`
	TotalChildCount int                  `json:"-"`
}

// Request Structure for Create Topic
type CreateTopicsRequest struct {
	Names  []string             `json:"names"`
	Topics []CreateTopicRequest `json:"topics"`
}

// Request Structure for Edit Topic
type EditTopicRequest struct {
	Name         string                 `json:"name"`
	IsEnabled    *bool                  `json:"is_enabled"`
	Priority     float32                `json:"priority"`
	IsSearchable *bool                  `json:"is_searchable"`
	Metadata     map[string]interface{} `json:"metadata"`
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
	TopicIds []string `json:"topic_ids" binding:"required"`
}
