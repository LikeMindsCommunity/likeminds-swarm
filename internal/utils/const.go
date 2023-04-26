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
)

const (
	CreatePostRoute    string = "route://create_post"
	HomeFeedRoute      string = "route://feed?type=universal"
	PostDetailRoute    string = "route://post_detail?post_id=%s"
	CommentDetailRoute string = "route://post_detail?post_id=%s&comment_id=%s"
)
