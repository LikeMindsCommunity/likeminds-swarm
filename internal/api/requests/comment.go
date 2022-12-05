package requests

type CreateCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

type DeleteCommentRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}
