package constants

type AttachmentType string

const (
	ImageType    AttachmentType = "image"
	VideoType    AttachmentType = "video"
	DocumentType AttachmentType = "document"
	LinkType     AttachmentType = "link"
)

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
