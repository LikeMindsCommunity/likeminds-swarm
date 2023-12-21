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
)

type PlatformType string

const DefaultCommunityId = -1

const SdkAuthenticateEndPoint = "/api/sdk/authenticate"
const SendNotificationEndPoint = "/api/external_service_apis/send_notifications"
const FetchMembersMetaEndPoint = "/api/community/fetch_members_meta"
const FetchCommunityConfigurations = "/api/community/configurations"

const ParamMemberIds = "member_ids"
const ParamCommunityId = "community_id"

const ContentTypeHeader = "application/json"
const SwarmServiceMemberIdHeader = "swarm-service"

const (
	PostCommunityConfigurationKey = "post"
)

const (
	FeedMetadataCommunityConfigurationType = "feed_metadata"
)

const (
	DefaultFeedMetadataPostVariableValue = "post"
)

const (
	CommunityConfigurationsCacheTTLInHours = 6
)
