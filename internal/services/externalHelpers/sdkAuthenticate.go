package externalHelpers

import (
	"github.com/gin-gonic/gin"
)

func GetCommunityId(c *gin.Context) int {
	//Send Request
	respBytes, statusCode, err := GetRequestResponse(CaravanService, SdkAuthenticateEndPoint, GETRequest, CreateHeaders(c, ""), nil, nil)
	if respBytes == nil {
		//If API fails or any other error
		GeneralAPIError(c, err.Error())
		return DefaultCommunityId
	}

	//Validate response
	apiCR := ValidateClientResponse(c, respBytes, statusCode)
	if apiCR == nil {
		return DefaultCommunityId
	}

	//If flow succeeds
	dataResponse := apiCR.Response

	return int(dataResponse["community_id"].(float64))
}
