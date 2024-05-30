package externalHelpers

const GETMethod = 0
const POSTMethod = 1
const PUTMethod = 2
const DELETEMethod = 3

type ServiceType string

const (
	CaravanService ServiceType = "caravan"
	KettleService  ServiceType = "kettle"
)

type RequestType int

const (
	GETRequest RequestType = iota
	POSTRequestRawBody
	POSTRequestFormUrlEncodedBody
	PUTRequest
	DELETERequest
	PATCHRequest
)

type PlatformType string

const DefaultCommunityId = -1

const SdkAuthenticateEndPoint = "/api/sdk/authenticate"
const SdkBotUserEndpoint = "/api/user/bot"
const SendNotificationEndPoint = "/api/external_service_apis/send_notifications"
const FetchMembersMetaEndPoint = "/api/community/fetch_members_meta"
const FetchCommunityConfigurationsEndpoint = "/api/community/configurations"
const FetchUserConnectionsEndPoint = "/api/community_member/%s/connection"
const SendMailEndpoint = "/api/external_service_apis/send_email"
const PushReportEndpoint = "/api/community/report"
const CommunityWebhooksEndpoint = "/api/webhook"
const KettleCacheDeleteEndpoint = "/cache"

const ParamMemberIds = "member_ids"
const ParamCommunityId = "community_id"
const ParamPage = "page"
const ParamPageSize = "page_size"

const ContentTypeHeader = "application/json"
const SwarmServiceHeader = "swarm-service"

const (
	PostCommunityConfigurationKey          = "post"
	CommentCommunityConfigurationKey       = "comment"
	UniversalFeedCommunityConfigurationKey = "universal_feed"
	LikeVariableCommunityConfigurationKey  = "like_entity_variable"
	LikeVariablePresentConfigurationKey    = "entity_name"
	LikeVariablePastConfigurationKey       = "past_tense_verb"
)

const (
	FeedMetadataCommunityConfigurationType   = "feed_metadata"
	NSFWFilteringCommunityConfigurationType  = "nsfw_filtering"
	PersonalisedFeedWeightsConfigurationType = "personalised_feed_weights"
)

const (
	DefaultPostVariableValue        = "post"
	DefaultCommentVariableValue     = "comment"
	DefaultLikePresentVariableValue = "like"
	DefaultLikePastVariableValue    = "liked"
)

const (
	FeedMetadataUniversalFeedCommentSortKey      = "comment_sort_order_key"
	FeedMetadataUniversalFeedCommentSortOrderKey = "comment_sort_order"
	FeedMetadataUniversalFeedCommentCountKey     = "comment_count"
)

// Inferdo API related constants
const (
	InferdoApiHeaderHost   = "nsfw-image-classification1.p.rapidapi.com"
	InferdoNsfwApiEndpoint = "https://nsfw-image-classification1.p.rapidapi.com/img/nsfw"
)

// Webhook Failure Mail Constants
const (
	WebhookFailureSubject = "Notification for webhook failure"
	WebhookFailureBody    = `Hey Team,<br>
	<b>%s</b> webhook has failed on %v. <br>
	Please inform the customer about the same.<br>

	<h2>Here are the details:</h2>
	<br>
	Webhook URL: %s<br>
	Webhook Failure Time: %v<br>
	Webhook Status Code: %d<br>
	Webhook Response: %v<br>
	Webhook Payload: <br>
	<code>
	%v
	</code>`
)
