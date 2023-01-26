package entities

type OGTags struct {
	Title       string `json:"title,omitempty" bson:"title,omitempty"`
	Image       string `json:"image,omitempty" bson:"image,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Url         string `json:"url,omitempty" bson:"url,omitempty"`
}

type AttachmentMeta struct {
	Url       string  `json:"url,omitempty" bson:"url,omitempty"`
	Format    string  `json:"format,omitempty" bson:"format,omitempty"`
	Size      int     `json:"size,omitempty" bson:"size,omitempty"`
	Duration  int     `json:"duration,omitempty" bson:"duration,omitempty"`
	PageCount int     `json:"page_count,omitempty" bson:"page_count,omitempty"`
	OgTags    *OGTags `json:"og_tags,omitempty" bson:"og_tags,omitempty"`
}

type Attachment struct {
	AttachmentType int             `json:"attachment_type" bson:"attachment_type"`
	AttachmentMeta *AttachmentMeta `json:"attachment_meta,omitempty" bson:"attachment_meta,omitempty"`
}

func NewAttachment(attachment_type int, attachment_meta AttachmentMeta) Attachment {
	return Attachment{
		AttachmentType: attachment_type,
		AttachmentMeta: &attachment_meta,
	}
}

func NewAttachmentMeta(url string, format string, size int, duration int, pageCount int, ogTags OGTags) AttachmentMeta {
	return AttachmentMeta{
		Url:       url,
		Format:    format,
		Size:      size,
		Duration:  duration,
		PageCount: pageCount,
		OgTags:    &ogTags,
	}
}

func NewOgTags(title string, image string, description string, url string) OGTags {
	return OGTags{
		Title:       title,
		Image:       image,
		Description: description,
		Url:         url,
	}
}
