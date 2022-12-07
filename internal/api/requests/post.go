package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

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

type AttachmentResponse struct {
	FileType   int    `json:"file_type"`
	FileUrl    string `json:"file_url,omitempty"`
	FileFormat string `json:"file_format,omitempty"`
	FileSize   string `json:"file_size,omitempty"`
}

type PostResponse struct {
	ID            primitive.ObjectID `json:"_id"`
	Text          string             `json:"text"`
	ApiKey        string             `json:"api_key"`
	IsPinned      bool               `json:"is_pinned"`
	UserId        string             `json:"user_id"`
	Attachments   []interface{}      `json:"attachments"`
	LikesCount    int                `json:"likes_count"`
	CommentsCount int                `json:"comments_count"`
	IsDeleted     bool               `json:"is_deleted,omitempty"`
	DeletedBy     string             `json:"deleted_by,omitempty"`
	DeleteReason  string             `json:"delete_reason,omitempty"`
	CreatedAt     int                `json:"created_at"`
	UpdatedAt     int                `json:"updated_at"`
}

type FetchPostResponse struct {
	PostResponse
	IsSaved bool              `json:"is_saved"`
	Replies []CommentResponse `json:"replies"`
}
