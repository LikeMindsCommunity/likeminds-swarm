package entities

import (
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Name                 string             `json:"name,omitempty" bson:"name,omitempty"`
	Url                  string             `json:"url,omitempty" bson:"url,omitempty"`
	Format               string             `json:"format,omitempty" bson:"format,omitempty"`
	Size                 int                `json:"size,omitempty" bson:"size,omitempty"`
	Duration             int                `json:"duration,omitempty" bson:"duration,omitempty"`
	PageCount            int                `json:"page_count,omitempty" bson:"page_count,omitempty"`
	ThumbnailUrl         string             `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	OgTags               *OGTags            `json:"og_tags,omitempty" bson:"og_tags,omitempty"`
	EntityID             primitive.ObjectID `json:"entity_id,omitempty" bson:"entity_id,omitempty"`
	CoverImageUrl        string             `json:"cover_image_url,omitempty" bson:"cover_image_url,omitempty"`
	Title                string             `json:"title,omitempty" bson:"title,omitempty"`
	Body                 string             `json:"body,omitempty" bson:"body,omitempty"`
	ExpiryTime           int64              `json:"expiry_time,omitempty" bson:"expiry_time,omitempty"`
	PollType             string             `json:"poll_type,omitempty" bson:"poll_type,omitempty"`
	MultipleSelectState  string             `json:"multiple_select_state,omitempty" bson:"multiple_select_state,omitempty"`
	MultipleSelectNumber int                `json:"multiple_select_number,omitempty" bson:"multiple_select_number,omitempty"`
	IsAnonymous          bool               `json:"is_anonymous,omitempty" bson:"is_anonymous,omitempty"`
	AllowAddOption       bool               `json:"allow_add_option,omitempty" bson:"allow_add_option,omitempty"`
	NsfwScore            float64            `json:"nsfw_score,omitempty" bson:"nsfw_score,omitempty"`
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
	ogTags OGTags, entityId primitive.ObjectID, coverImageUrl string, title string, body string, expiryTime int64, pollType string,
	multipleSelectState string, multipleSelectNumber int, isAnonymous bool, allowAddOption bool, nsfwScore float64) AttachmentMeta {

	attachmentMeta := AttachmentMeta{
		Name:                 name,
		Url:                  url,
		Format:               format,
		Size:                 size,
		Duration:             duration,
		PageCount:            pageCount,
		OgTags:               &ogTags,
		ThumbnailUrl:         thumbnailUrl,
		CoverImageUrl:        coverImageUrl,
		Title:                title,
		Body:                 body,
		ExpiryTime:           expiryTime,
		PollType:             pollType,
		MultipleSelectState:  multipleSelectState,
		MultipleSelectNumber: multipleSelectNumber,
		IsAnonymous:          isAnonymous,
		AllowAddOption:       allowAddOption,
	}

	if entityId != primitive.NilObjectID {
		attachmentMeta.EntityID = entityId
	}

	if nsfwScore != 0.0 {
		attachmentMeta.NsfwScore = nsfwScore
	}

	return attachmentMeta
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
