package responses

// Response Structure for Topic
type TopicResponse struct {
	ID           string  `json:"_id"`
	Name         string  `json:"name"`
	IsEnabled    bool    `json:"is_enabled"`
	Priority     float32 `json:"priority"`
	IsSearchable bool    `json:"is_searchable"`
	ParentId     string  `json:"parent_id"`
	ParentName   string  `json:"parent_name"`
	Level        int     `json:"level"`
	WidgetId     string  `json:"widget_id"`
}

// Response Structure for topics with meta
type TopicResponseWithMeta struct {
	*TopicResponse
	NumberOfPosts   int `json:"number_of_posts"`
	TotalChildCount int `json:"total_child_count"`
}
