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

const (
	PlatformAndroid string = "an"
	PlatformWeb     string = "web"
	PlatformIoS     string = "ios"
)

const DefaultCommunityId = -1

const SdkAuthenticateEndPoint = "/api/sdk/authenticate"
const SendNotificationEndPoint = "/api/external_service_apis/send_notifications"
