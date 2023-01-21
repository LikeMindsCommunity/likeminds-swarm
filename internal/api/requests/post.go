package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Attachment struct {
	FileType   int    `json:"file_type" binding:"required"`
	FileUrl    string `json:"file_url"`
	FileFormat string `json:"file_format"`
	FileSize   string `json:"file_size"`
}

type CreatePostRequest struct {
	Text        string       `json:"text" binding:"required"`
	Attachments []Attachment `json:"attachments"`
	ChatroomID  int          `json:"feedroom_id"`
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
	Communityid   int                `json:"community_id"`
	IsPinned      bool               `json:"is_pinned"`
	UserId        string             `json:"user_id"`
	Attachments   []entities.Widget  `json:"attachments"`
	LikesCount    int                `json:"likes_count"`
	CommentsCount int                `json:"comments_count"`
	IsDeleted     bool               `json:"is_deleted,omitempty"`
	DeletedBy     string             `json:"deleted_by,omitempty"`
	DeleteReason  string             `json:"delete_reason,omitempty"`
	IsSaved       bool               `json:"is_saved"`
	MenuItems     []MenuResponse     `json:"menu_items"`
	CreatedAt     int                `json:"created_at"`
	UpdatedAt     int                `json:"updated_at"`
}

type MenuResponse struct {
	Title string `json:"title"`
	Route string `json:"route,omitempty"`
}

type FetchPostResponse struct {
	PostResponse
	Replies []CommentResponse `json:"replies"`
}

type FetchUserMultiplePostResponse struct {
	Success    bool           `json:"success"`
	TotalCount int            `json:"total_count,omitempty"`
	Posts      []PostResponse `json:"posts"`
}
