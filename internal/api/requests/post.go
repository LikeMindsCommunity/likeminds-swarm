package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/entities"
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
	Text              string       `json:"text"`
	Heading           string       `json:"heading"`
	TempID            *string      `json:"temp_id"`
	TopicIds          []string     `json:"topic_ids"`
	Attachments       []Attachment `json:"attachments"`
	ChatroomID        int          `json:"feedroom_id"`
	UUIDs             []string     `json:"uuids"`
	On_behalf_of_uuid string       `json:"on_behalf_of_uuid,omitempty"`
	Visibility        string       `json:"visibility"`
	User_is_cm        bool         `json:"user_is_cm,omitempty"`
	IsRepost          bool         `json:"is_repost"`
	CreatedAt         int          `json:"created_at"`
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

// Resonse Structure for Post
type PostResponse struct {
	ID                 primitive.ObjectID    `json:"_id"`
	TempID             *string               `json:"temp_id"`
	Topics             []primitive.ObjectID  `json:"topics"`
	Text               string                `json:"text"`
	Heading            string                `json:"heading"`
	CommunityId        int                   `json:"community_id,omitempty"`
	ChatroomId         int                   `json:"feedroom_id,omitempty"`
	IsPinned           bool                  `json:"is_pinned"`
	UserId             string                `json:"user_id,omitempty"`
	UUID               string                `json:"uuid,omitempty"`
	Attachments        []entities.Attachment `json:"attachments"`
	LikesCount         int                   `json:"likes_count"`
	CommentsCount      int                   `json:"comments_count"`
	RepostCount        int32                 `json:"repost_count"`
	IsDeleted          bool                  `json:"is_deleted,omitempty"`
	IsEdited           bool                  `json:"is_edited"`
	IsRepost           bool                  `json:"is_repost"`
	IsRepostedByUser   bool                  `json:"is_reposted_by_user"`
	OriginalAuthorUUID string                `json:"original_author_uuid,omitempty"`
	DeletedBy          string                `json:"deleted_by,omitempty"`
	DeletedByUUID      string                `json:"deleted_by_uuid,omitempty"`
	DeleteReason       string                `json:"delete_reason,omitempty"`
	IsLiked            bool                  `json:"is_liked"`
	IsSaved            bool                  `json:"is_saved"`
	MenuItems          []MenuResponse        `json:"menu_items"`
	CreatedAt          int                   `json:"created_at"`
	UpdatedAt          int                   `json:"updated_at"`
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
