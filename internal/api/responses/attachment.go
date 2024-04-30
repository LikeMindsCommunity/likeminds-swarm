package responses

import (
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
)

// Response Structure for OG Tags
type OGTags struct {
	Title       string `json:"title,omitempty"`
	Image       string `json:"image,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

// Response Structure for Attachment Meta
type AttachmentMeta struct {
	Name                 string  `json:"name,omitempty"`
	Url                  string  `json:"url,omitempty"`
	Format               string  `json:"format,omitempty"`
	Size                 int     `json:"size,omitempty"`
	Duration             int     `json:"duration,omitempty"`
	PageCount            int     `json:"page_count,omitempty"`
	ThumbnailUrl         string  `json:"thumbnail_url,omitempty"`
	OgTags               *OGTags `json:"og_tags,omitempty"`
	EntityID             string  `json:"entity_id,omitempty"`
	CoverImageUrl        string  `json:"cover_image_url,omitempty"`
	Title                string  `json:"title,omitempty"`
	Body                 string  `json:"body,omitempty"`
	ExpiryTime           int64   `json:"expiry_time,omitempty"`
	PollType             string  `json:"poll_type,omitempty"`
	MultipleSelectState  string  `json:"multiple_select_state,omitempty"`
	MultipleSelectNumber int     `json:"multiple_select_number,omitempty"`
	IsAnonymous          bool    `json:"is_anonymous,omitempty"`
	AllowAddOption       bool    `json:"allow_add_option,omitempty"`
	NsfwScore            float64 `json:"nsfw_score,omitempty"`
}

// Response Structure for Attachment
type Attachment struct {
	AttachmentType int                  `json:"attachment_type,omitempty"`
	AttachmentMeta *AttachmentMeta      `json:"attachment_meta,omitempty"`
	Type           enums.AttachmentType `json:"type,omitempty"`
	MetaData       *AttachmentMeta      `json:"meta_data,omitempty"`
}
