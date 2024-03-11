package utils

import (
	"github.com/gin-gonic/gin"
)

// GetHeaders Used to get headers from API request
func GetHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	headers[HeadersMemberId] = c.GetHeader(HeadersMemberId)
	headers[HeadersPlatformCode] = c.GetHeader(HeadersPlatformCode)
	headers[HeadersVersionCode] = c.GetHeader(HeadersVersionCode)
	headers[HeadersSdkSource] = c.GetHeader(HeadersSdkSource)
	headers[HeadersDeviceId] = c.GetHeader(HeadersDeviceId)
	headers[HeadersApiKey] = c.GetHeader(HeadersApiKey)
	headers[HeadersAcceptVersion] = c.GetHeader(HeadersAcceptVersion)
	headers[HeaderMemberRole] = c.GetHeader(HeaderMemberRole)

	return headers
}
