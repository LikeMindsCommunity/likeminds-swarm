package enums

// constants for attachment_type
const (
	ImageWidget int = iota + 1
	VideoWidget
	DocumentWidget
	LinkWidget
)

// AttachmentType represents the type of attachment
type AttachmentType string

// enum values for attachment TYPE
const (
	ImageType    AttachmentType = "image"
	VideoType    AttachmentType = "video"
	DocumentType AttachmentType = "document"
	LinkType     AttachmentType = "link"
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
	}

	return ""
}

// checks if the attachment type is valid
func (at AttachmentType) IsValid() bool {
	switch at {
	case ImageType, VideoType, DocumentType, LinkType:
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
	}
	return 0
}
