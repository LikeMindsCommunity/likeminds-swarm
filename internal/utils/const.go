package utils

const (
	HeadersMemberId      string = "x-member-id"
	HeadersVersionCode   string = "x-version-code"
	HeadersPlatformCode  string = "x-platform-code"
	HeadersPlatformType  string = "x-platform-type"
	HeadersSdkSource     string = "x-sdk-source"
	HeadersDeviceId      string = "x-device-id"
	HeadersApiKey        string = "x-api-key"
	HeadersAcceptVersion string = "x-accept-version"
	HeadersApiVersion    string = "x-api-version"
	HeaderMemberRole     string = "x-member-role"
	HeaderContentType    string = "Content-Type"
)

const (
	ContentTypeApplicationJson string = "application/json"
)

const (
	SdkSourceFeed string = "feed"
)

const (
	CreatePostRoute        string = "route://create_post"
	HomeFeedRoute          string = "route://feed?type=universal"
	PostDetailRoute        string = "route://post_detail?post_id=%s"
	CommentDetailRoute     string = "route://post_detail?post_id=%s&comment_id=%s"
	PendingPostDetailRoute string = "route://pending_post_detail?post_id=%s"
)

// Member Roles
const (
	GuestRole   string = "GUEST"
	CMRole      string = "CM"
	DefaultRole string = ""
)

// checks whether the member is CM or not based on member role received in header
func IsCMRole(memberRole string) bool {
	return memberRole == CMRole
}
