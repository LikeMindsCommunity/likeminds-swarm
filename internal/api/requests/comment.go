package requests

// Request Structure for Create Comment
type CreateCommentRequest struct {
	Text        string       `json:"text" binding:"required"`
	Attachments []Attachment `json:"attachments"`
	TempID      *string      `json:"temp_id"`
	UUIDs       []string     `json:"uuids"`
	CreatedAt   int          `json:"created_at"`
}

// Request Structure for Edit Comment
type EditCommentRequest struct {
	Text     string `json:"text" binding:"required"`
	UserIsCm bool   `json:"user_is_cm"`
}

// Request Structure for Delete Comment
type DeleteCommentRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}
