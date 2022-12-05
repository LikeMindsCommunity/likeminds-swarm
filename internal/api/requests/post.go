package requests

type Attachment struct {
	FileType   int    `json:"file_type" binding:"required"`
	FileUrl    string `json:"file_url"`
	FileFormat string `json:"file_format"`
	FileSize   string `json:"file_size"`
}

type CreatePostRequest struct {
	Text        string       `json:"text" binding:"required"`
	Attachments []Attachment `json:"attachments"`
}

type DeletePostRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}
