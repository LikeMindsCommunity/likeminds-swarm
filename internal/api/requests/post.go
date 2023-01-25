package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OGTags struct {
	Title       string `json:"title"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

type AttachmentMeta struct {
	Url       string `json:"url"`
	Format    string `json:"format"`
	Size      int    `json:"size"`
	Duration  int    `json:"duration"`
	PageCount int    `json:"page_count"`
	OgTags    OGTags `json:"og_tags"`
}

type Attachment struct {
	AttachmentType int            `json:"attachment_type" binding:"required"`
	AttachmentMeta AttachmentMeta `json:"attachment_meta"`
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

type PostResponse struct {
	ID            primitive.ObjectID    `json:"_id"`
	Text          string                `json:"text"`
	CommunityId   int                   `json:"community_id"`
	ChatroomId    int                   `json:"feedroom_id,omitempty"`
	IsPinned      bool                  `json:"is_pinned"`
	UserId        string                `json:"user_id"`
	Attachments   []entities.Attachment `json:"attachments"`
	LikesCount    int                   `json:"likes_count"`
	CommentsCount int                   `json:"comments_count"`
	IsDeleted     bool                  `json:"is_deleted,omitempty"`
	DeletedBy     string                `json:"deleted_by,omitempty"`
	DeleteReason  string                `json:"delete_reason,omitempty"`
	IsLiked       bool                  `json:"is_liked"`
	IsSaved       bool                  `json:"is_saved"`
	MenuItems     []MenuResponse        `json:"menu_items"`
	CreatedAt     int                   `json:"created_at"`
	UpdatedAt     int                   `json:"updated_at"`
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
