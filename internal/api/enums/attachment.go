package enums

import "github.com/nateshr/likeminds-swarm/internal/api/constants"

type AttachmentType string

const (
	ImageType    AttachmentType = "image"
	VideoType    AttachmentType = "video"
	DocumentType AttachmentType = "document"
	LinkType     AttachmentType = "link"
)

// Create New Attachment Type from int
func NewAttachmentTypeFromInt(attachment_type int) AttachmentType {
	switch attachment_type {
	case constants.ImageWidget:
		return ImageType
	case constants.VideoWidget:
		return VideoType
	case constants.DocumentWidget:
		return DocumentType
	case constants.LinkWidget:
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
		return constants.ImageWidget
	case VideoType:
		return constants.VideoWidget
	case DocumentType:
		return constants.DocumentWidget
	case LinkType:
		return constants.LinkWidget
	}
	return 0
}
