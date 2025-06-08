package enums

// constants for attachment_type
const (
	NoAttachment int = iota
	ImageWidget
	VideoWidget
	DocumentWidget
	LinkWidget
	CustomWidget
	PollWidget
	ArticleWidget
	PostWidget
	RepostWidget
	GIFWidget
	ReelWidget
)

// AttachmentType represents the type of attachment
type AttachmentType string

// enum values for attachment TYPE
const (
	ImageType    AttachmentType = "image"
	VideoType    AttachmentType = "video"
	DocumentType AttachmentType = "document"
	LinkType     AttachmentType = "link"
	CustomType   AttachmentType = "custom"
	PollType     AttachmentType = "poll"
	ArticleType  AttachmentType = "article"
	PostType     AttachmentType = "post"
	RepostType   AttachmentType = "repost"
	GIFType      AttachmentType = "gif"
	ReelType     AttachmentType = "reel"
)

// Create New Attachment Type from int
func NewAttachmentTypeFromInt(attachment_type int) AttachmentType {
	switch attachment_type {
	case ImageWidget:
		return ImageType
	case VideoWidget:
		return VideoType
	case DocumentWidget:
		return DocumentType
	case LinkWidget:
		return LinkType
	case CustomWidget:
		return CustomType
	case PollWidget:
		return PollType
	case ArticleWidget:
		return ArticleType
	case PostWidget:
		return PostType
	case RepostWidget:
		return RepostType
	case GIFWidget:
		return GIFType
	case ReelWidget:
		return ReelType
	}
	return ""
}

// checks if the attachment type is valid
func (at AttachmentType) IsValid() bool {
	switch at {
	case ImageType, VideoType, DocumentType, LinkType, CustomType, PollType, ArticleType, PostType, RepostType, GIFType, ReelType:
		return true
	}
	return false
}

// IsValidRepostAttachment | checks if the attachment type is valid repost attachment
func (at AttachmentType) IsValidRepostAttachment() bool {
	switch at {
	case PostType:
		return true
	default:
		return false
	}
}

// function to convert attachment type to string
func (at AttachmentType) ToString() string {
	return string(at)
}

// function to convert attachment type to int
func (at AttachmentType) ToInt() int {
	switch at {
	case ImageType:
		return ImageWidget
	case VideoType:
		return VideoWidget
	case DocumentType:
		return DocumentWidget
	case LinkType:
		return LinkWidget
	case CustomType:
		return CustomWidget
	case PollType:
		return PollWidget
	case ArticleType:
		return ArticleWidget
	case PostType:
		return PostWidget
	case RepostType:
		return RepostWidget
	case GIFType:
		return GIFWidget
	case ReelType:
		return ReelWidget
	}
	return 0
}
