package enums

// constants for attachment_type
const (
	ImageWidget int = iota + 1
	VideoWidget
	DocumentWidget
	LinkWidget
	CustomWidget
	PollWidget
	ArticleWidget
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
	}

	return ""
}

// checks if the attachment type is valid
func (at AttachmentType) IsValid() bool {
	switch at {
	case ImageType, VideoType, DocumentType, LinkType,
		CustomType, PollType, ArticleType:
		return true
	}
	return false
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
	}
	return 0
}

// constants for poll_type
const (
	InstantPollType  string = "instant"
	DeferredPollType string = "deferred"
)

// function to check if poll type is valid
func IsPollTypeValid(pollType string) bool {
	switch pollType {
	case InstantPollType, DeferredPollType:
		return true
	}
	return false
}

// constants for multiple_select_state in poll
const (
	ExactlySelectStateType string = "exactly"
	AtMaxSelectStateType   string = "at_max"
	AtLeastSelectStateType string = "at_least"
)

// function to check if poll multiple select state is valid
func IsPollMultipleSelectStateValid(pollType string) bool {
	switch pollType {
	case ExactlySelectStateType, AtLeastSelectStateType,
		AtMaxSelectStateType:
		return true
	}
	return false
}
