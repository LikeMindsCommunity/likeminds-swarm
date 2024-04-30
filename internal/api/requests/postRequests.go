package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OG Tags Structure
type OGTags struct {
	Title       string `json:"title,omitempty"`
	Image       string `json:"image,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

// Attachment Meta Structure
type AttachmentMeta struct {
	Name                 string                 `json:"name,omitempty"`
	Url                  string                 `json:"url,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Size                 int                    `json:"size,omitempty"`
	Duration             int                    `json:"duration,omitempty"`
	PageCount            int                    `json:"page_count,omitempty"`
	ThumbnailUrl         string                 `json:"thumbnail_url,omitempty"`
	OgTags               OGTags                 `json:"og_tags,omitempty"`
	EntityID             string                 `json:"entity_id,omitempty"`
	CoverImageUrl        string                 `json:"cover_image_url,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Body                 string                 `json:"body,omitempty"`
	Options              []string               `json:"options,omitempty"`
	ExpiryTime           int64                  `json:"expiry_time,omitempty"`
	PollType             string                 `json:"poll_type,omitempty"`
	MultipleSelectState  string                 `json:"multiple_select_state,omitempty"`
	MultipleSelectNumber int                    `json:"multiple_select_number,omitempty"`
	IsAnonymous          bool                   `json:"is_anonymous,omitempty"`
	AllowAddOption       bool                   `json:"allow_add_option,omitempty"`
	PostID               string                 `json:"post_id,omitempty"`
	WidgetMeta           map[string]interface{} `json:"widget_meta,omitempty"`
	NsfwScore            float64                `json:"-,omitempty"` // field to be updated internally
}

// Attachment Structure
type Attachment struct {
	AttachmentType int                  `json:"attachment_type"`
	AttachmentMeta AttachmentMeta       `json:"attachment_meta"`
	Type           enums.AttachmentType `json:"type"`
	MetaData       AttachmentMeta       `json:"meta_data"`
}

// Request Structure for Create Post
type CreatePostRequest struct {
	Text           string               `json:"text"`
	Heading        string               `json:"heading"`
	TempID         *string              `json:"temp_id"`
	TopicIds       []string             `json:"topic_ids"`
	Attachments    []Attachment         `json:"attachments"`
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
	Text        string       `json:"text"`
	Heading     string       `json:"heading"`
	TopicIds    []string     `json:"topic_ids,omitempty"`
	Attachments []Attachment `json:"attachments"`
	Visibility  string       `json:"visibility"`
	UserIsCm    bool         `json:"user_is_cm"`
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
