package requests

import "github.com/nateshr/likeminds-swarm/internal/api/enums"

// OG Tags Structure
type OGTagsRequest struct {
	Title       string `json:"title,omitempty"`
	Image       string `json:"image,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

// Attachment Meta Structure
type AttachmentMetaRequest struct {
	Name                 string                 `json:"name,omitempty"`
	Url                  string                 `json:"url,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Size                 int                    `json:"size,omitempty"`
	Duration             int                    `json:"duration,omitempty"`
	Width                int                    `json:"width,omitempty"`
	Height               int                    `json:"height,omitempty"`
	PageCount            int                    `json:"page_count,omitempty"`
	ThumbnailUrl         string                 `json:"thumbnail_url,omitempty"`
	OgTags               OGTagsRequest          `json:"og_tags,omitempty"`
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

// AttachmentRequest Structure
type AttachmentRequest struct {
	AttachmentType int                   `json:"attachment_type"`
	AttachmentMeta AttachmentMetaRequest `json:"attachment_meta"`
	Type           enums.AttachmentType  `json:"type"`
	MetaData       AttachmentMetaRequest `json:"meta_data"`
}
