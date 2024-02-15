package externalHelpers

const GETMethod = 0
const POSTMethod = 1
const PUTMethod = 2
const DELETEMethod = 3

type ServiceType int

const (
	CaravanService ServiceType = iota
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
const CommunityWebhooksEndpoint = "/api/community/webhook"

const ParamMemberIds = "member_ids"
const ParamCommunityId = "community_id"
const ParamPage = "page"
const ParamPageSize = "page_size"

const ContentTypeHeader = "application/json"
const SwarmServiceHeader = "swarm-service"

const (
	PostCommunityConfigurationKey = "post"
)

const (
	FeedMetadataCommunityConfigurationType  = "feed_metadata"
	NSFWFilteringCommunityConfigurationType = "nsfw_filtering"
)

const (
	DefaultFeedMetadataPostVariableValue = "post"
)

const (
	CommunityConfigurationsCacheTTLInHours = 175 // 7 days
	CommunityWebhooksCacheTTTLInHours      = 175 // 7 days
)

// Inferdo API related constants
const (
	InferdoApiHeaderHost   = "nsfw-image-classification1.p.rapidapi.com"
	InferdoNsfwApiEndpoint = "https://nsfw-image-classification1.p.rapidapi.com/img/nsfw"
)
