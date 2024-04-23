package requests

// struct to validate request to fetch user topics
type FetchUserTopicsRequest struct {
	UUIDs string `form:"uuids" binding:"required"`
}

// struct to validate request to update user topics
type UpdateUserTopicsRequest struct {
	TopicIds map[string]bool `json:"topic_ids" binding:"required"`
}
