package enums

const (
	UniversalFeedTopLikedComments = "likes"
)

const (
	AscendingSortOrder  = "asc"
	DescendingSortOrder = "desc"
)

type FeedType string

const (
	SocialFeedType FeedType = "social_feed"
	QnAFeedType    FeedType = "qna_feed"
	VideoFeedType  FeedType = "video_feed"
)

// function to get attchments list from feed type
func (ft FeedType) GetAttachmentsList() []int {
	switch ft {
	case SocialFeedType:
		return []int{ImageWidget, VideoWidget, DocumentWidget, LinkWidget, CustomWidget, PollWidget, ArticleWidget, PostWidget, RepostWidget, GIFWidget, -1} // -1 is for no attachments
	case QnAFeedType:
		return []int{ImageWidget, VideoWidget, DocumentWidget, LinkWidget, CustomWidget, PollWidget, ArticleWidget, PostWidget, RepostWidget, GIFWidget, -1} // -1 is for no attachments
	case VideoFeedType:
		return []int{ReelWidget}
	}
	return []int{}
}
