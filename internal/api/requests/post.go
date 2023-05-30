package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OG Tags Structure
type OGTags struct {
	Title       string `json:"title"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

// Attachment Meta Structure
type AttachmentMeta struct {
	Name         string `json:"name"`
	Url          string `json:"url"`
	Format       string `json:"format"`
	Size         int    `json:"size"`
	Duration     int    `json:"duration"`
	PageCount    int    `json:"page_count"`
	ThumbnailUrl string `json:"thumbnail_url"`
	OgTags       OGTags `json:"og_tags"`
}

// Attachment Structure
type Attachment struct {
	AttachmentType int            `json:"attachment_type" binding:"required"`
	AttachmentMeta AttachmentMeta `json:"attachment_meta"`
}

// Request Structure for Create Post
type CreatePostRequest struct {
	Text        string       `json:"text"`
	Heading     string       `json:"heading"`
	Attachments []Attachment `json:"attachments"`
	ChatroomID  int          `json:"feedroom_id"`
	UUIDs       []string     `json:"uuids"`
}

// Request Structure for Edit Post
type EditPostRequest struct {
	Text        string       `json:"text"`
	Heading     string       `json:"heading"`
	Attachments []Attachment `json:"attachments"`
	UserIsCm    bool         `json:"user_is_cm"`
}

// Request Structure for Delete Post
type DeletePostRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}

// Resonse Structure for Post
type PostResponse struct {
	ID            primitive.ObjectID    `json:"_id"`
	Text          string                `json:"text"`
	Heading       string                `json:"heading"`
	CommunityId   int                   `json:"community_id"`
	ChatroomId    int                   `json:"feedroom_id,omitempty"`
	IsPinned      bool                  `json:"is_pinned"`
	UserId        string                `json:"user_id"`
	Attachments   []entities.Attachment `json:"attachments"`
	LikesCount    int                   `json:"likes_count"`
	CommentsCount int                   `json:"comments_count"`
	IsDeleted     bool                  `json:"is_deleted,omitempty"`
	IsEdited      bool                  `json:"is_edited"`
	DeletedBy     string                `json:"deleted_by,omitempty"`
	DeleteReason  string                `json:"delete_reason,omitempty"`
	IsLiked       bool                  `json:"is_liked"`
	IsSaved       bool                  `json:"is_saved"`
	MenuItems     []MenuResponse        `json:"menu_items"`
	CreatedAt     int                   `json:"created_at"`
	UpdatedAt     int                   `json:"updated_at"`
}

// Response Structure for Menu Item of an Entity
type MenuResponse struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Route string `json:"route,omitempty"`
}

// Response Structure for Fetch Post
type FetchPostResponse struct {
	PostResponse
	Replies []CommentResponse `json:"replies"`
}

// Response Structure for Fetch Multiple Post
type FetchUserMultiplePostResponse struct {
	Success    bool           `json:"success"`
	TotalCount int            `json:"total_count,omitempty"`
	Posts      []PostResponse `json:"posts"`
}

// Request Structure for Search Post
type SearchPostRequest struct {
	Search              string `form:"search"`
	SearchType          string `form:"search_type"`
	ExcludedChatroomIDs string `form:"excluded_chatroom_ids"`
	UserIsCm            bool   `form:"user_is_cm"`
}
