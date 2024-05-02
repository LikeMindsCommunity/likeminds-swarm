package requests

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Request Structure for Create Post
type CreatePostRequest struct {
	Text           string               `json:"text"`
	Heading        string               `json:"heading"`
	TempID         *string              `json:"temp_id"`
	TopicIds       []string             `json:"topic_ids"`
	Attachments    []AttachmentRequest  `json:"attachments"`
	ChatroomID     int                  `json:"feedroom_id"`
	UUIDs          []string             `json:"uuids"`
	OnBehalfOfUUID string               `json:"on_behalf_of_uuid,omitempty"`
	Visibility     string               `json:"visibility"`
	UserIsCm       bool                 `json:"user_is_cm,omitempty"`
	IsRepost       bool                 `json:"is_repost"`
	CreatedAt      int                  `json:"created_at"`
	ParsedTopicIds []primitive.ObjectID `json:"-"` // field to be updated internally
	OriginalAuthor string               `json:"-"` // field to be updated internally
	PostType       string               `json:"-"` // field to be updated internally
}

// Request Structure for Edit Post
type EditPostRequest struct {
	Text        string              `json:"text"`
	Heading     string              `json:"heading"`
	TopicIds    []string            `json:"topic_ids,omitempty"`
	Attachments []AttachmentRequest `json:"attachments"`
	Visibility  string              `json:"visibility"`
	UserIsCm    bool                `json:"user_is_cm"`
}

// Request Structure for Delete Post
type DeletePostRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}

// Request Structure for Search Post
type SearchPostRequest struct {
	Search              string `form:"search"`
	SearchType          string `form:"search_type"`
	ExcludedChatroomIDs string `form:"excluded_chatroom_ids"`
	UserIsCm            bool   `form:"user_is_cm"`
}

// Query Structure for fetch posts
type FetchPostsQueryRequest struct {
	PostIds        string `form:"post_ids"`
	PendingPostIds string `form:"pending_post_ids"`
	UserIsCm       bool   `form:"user_is_cm"`
}
