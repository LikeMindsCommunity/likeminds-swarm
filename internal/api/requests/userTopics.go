package requests

type UpdateUserTopicsRequest struct {
	TopicIds map[string]bool `json:"topic_ids" binding:"required"`
	UserIsCM bool            `json:"user_is_cm"`
}
