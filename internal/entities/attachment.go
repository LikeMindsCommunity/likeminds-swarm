package entities

import (
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
)

// Structure for OG Tags
type OGTags struct {
	Title       string `json:"title,omitempty" bson:"title,omitempty"`
	Image       string `json:"image,omitempty" bson:"image,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Url         string `json:"url,omitempty" bson:"url,omitempty"`
}

// Structure for Attachment Meta
type AttachmentMeta struct {
	Name         string  `json:"name,omitempty" bson:"name,omitempty"`
	Url          string  `json:"url,omitempty" bson:"url,omitempty"`
	Format       string  `json:"format,omitempty" bson:"format,omitempty"`
	Size         int     `json:"size,omitempty" bson:"size,omitempty"`
	Duration     int     `json:"duration,omitempty" bson:"duration,omitempty"`
	PageCount    int     `json:"page_count,omitempty" bson:"page_count,omitempty"`
	ThumbnailUrl string  `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	OgTags       *OGTags `json:"og_tags,omitempty" bson:"og_tags,omitempty"`
}

// Structure for Attachment
type Attachment struct {
	AttachmentType int                  `json:"attachment_type,omitempty" bson:"attachment_type,omitempty"`
	AttachmentMeta *AttachmentMeta      `json:"attachment_meta,omitempty" bson:"attachment_meta,omitempty"`
	Type           enums.AttachmentType `json:"type,omitempty" bson:"type,omitempty"`
	MetaData       *AttachmentMeta      `json:"meta_data,omitempty" bson:"meta_data,omitempty"`
}

// Exposed Method to Create New Attachment
func NewAttachment(attachment_type int, attachment_meta AttachmentMeta) Attachment {
	return Attachment{
		AttachmentType: attachment_type,
		AttachmentMeta: &attachment_meta,
	}
}

// Exposed Method to Create New Attachment Meta
func NewAttachmentMeta(name string, url string, format string, size int, duration int, pageCount int, thumbnailUrl string,
	ogTags OGTags) AttachmentMeta {
	return AttachmentMeta{
		Name:         name,
		Url:          url,
		Format:       format,
		Size:         size,
		Duration:     duration,
		PageCount:    pageCount,
		ThumbnailUrl: thumbnailUrl,
		OgTags:       &ogTags,
	}
}

// Exposed Method to Create New Og Tags
func NewOgTags(title string, image string, description string, url string) OGTags {
	return OGTags{
		Title:       title,
		Image:       image,
		Description: description,
		Url:         url,
	}
}
