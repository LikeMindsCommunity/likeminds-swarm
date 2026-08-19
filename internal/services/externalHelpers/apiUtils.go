package externalHelpers

import (
	"encoding/json"
	"net/http"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/utils"
	"github.com/gin-gonic/gin"
)

// Structure for Response for all API calls
type Response struct {
	Success      bool        `json:"success"`
	ErrorMessage string      `json:"error_message,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

// APIClientResponse Used only for internal API calls
type APIClientResponse struct {
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message"`
	Response     map[string]interface{} `json:"-"`
}

// Exposed Method to send General API Error
func GeneralAPIError(c *gin.Context, errorMessage string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success:      false,
		ErrorMessage: errorMessage,
	})
}

// UnmarshalAPIClientResponse used to unmarshal APIClientResponse i.e internal API call response
func UnmarshalAPIClientResponse(resp []byte, apiCR *APIClientResponse) error {
	if err := json.Unmarshal(resp, &apiCR); err != nil {
		return err
	}

	if err := json.Unmarshal(resp, &apiCR.Response); err != nil {
		return err
	}
	delete(apiCR.Response, "success")
	delete(apiCR.Response, "error_message")
	return nil
}

// CreateHeaders Used to create headers for our internal APIs
func CreateHeaders(c *gin.Context, userUniqueID string) map[string]interface{} {
	headers := make(map[string]interface{})
	if len(userUniqueID) > 0 {
		headers[utils.HeadersMemberId] = userUniqueID
	}
	headers[utils.HeadersPlatformCode] = c.GetHeader(utils.HeadersPlatformCode)
	headers[utils.HeadersVersionCode] = c.GetHeader(utils.HeadersVersionCode)
	headers[utils.HeadersSdkSource] = c.GetHeader(utils.HeadersSdkSource)
	headers[utils.HeadersDeviceId] = c.GetHeader(utils.HeadersDeviceId)
	headers[utils.HeadersApiKey] = c.GetHeader(utils.HeadersApiKey)
	headers[utils.HeadersAcceptVersion] = c.GetHeader(utils.HeadersAcceptVersion)
	return headers
}

// Exposed Method to fetch an http Request response
func GetRequestResponse(serviceType ServiceType, url string, requestType RequestType, headers map[string]interface{}, params map[string]string, body interface{}) ([]byte, int, error) {
	//Create internal API client
	client := NewAPIClient()
	var baseUrl string
	var respBytes []byte
	var statusCode int
	var err error

	switch serviceType {
	case CaravanService:
		baseUrl = client.CaravanServiceBaseURL
	case KettleService:
		baseUrl = client.KettleServiceBaseURL
	}

	switch requestType {
	case GETRequest:

		options := GetRequestOptions{
			Url:           baseUrl + url,
			CustomHeaders: headers,
			Params:        params,
		}

		respBytes, statusCode, err = client.GetRequest(&options)

	case POSTRequestRawBody:

		options := PostRequestOptions{
			Url:           baseUrl + url,
			CustomHeaders: headers,
			Params:        params,
			Body:          body,
		}

		respBytes, statusCode, err = client.PostRequest(&options, BodyTypeRaw)

	case POSTRequestFormUrlEncodedBody:

		options := PostRequestOptions{
			Url:           baseUrl + url,
			CustomHeaders: headers,
			Params:        params,
			Body:          body,
		}

		respBytes, statusCode, err = client.PostRequest(&options, BodyTypeFormUrlEncoded)

	case PUTRequest:

		options := PostRequestOptions{
			Url:           baseUrl + url,
			CustomHeaders: headers,
			Params:        params,
			Body:          body,
		}

		respBytes, statusCode, err = client.PutRequest(&options)

	case DELETERequest:

		options := PostRequestOptions{
			Url:           baseUrl + url,
			CustomHeaders: headers,
			Params:        params,
			Body:          body,
		}

		respBytes, statusCode, err = client.DeleteRequest(&options)

	case PATCHRequest:

		options := PostRequestOptions{
			Url:           baseUrl + url,
			CustomHeaders: headers,
			Params:        params,
			Body:          body,
		}

		respBytes, statusCode, err = client.PatchRequest(&options)
	}

	if err != nil {
		return nil, DefaultStatusCode, err
	}

	return respBytes, statusCode, nil
}

// Exposed Method to Validate response of a http Request
func ValidateClientResponse(c *gin.Context, respBytes []byte, statusCode int) *APIClientResponse {
	//Parse response
	var apiCR APIClientResponse
	err := UnmarshalAPIClientResponse(respBytes, &apiCR)

	if err != nil {
		//Internal unmarshal error
		GeneralAPIError(c, err.Error())
		return nil
	}

	if !apiCR.Success {
		//If internal api returns success as false
		c.JSON(statusCode, apiCR)
		return nil
	}

	return &apiCR
}

// Generate Response to be sent on request success
func GenerateResponse(c *gin.Context, dataResponse map[string]interface{}) {
	//Generating Response Object
	response := Response{
		Success: true,
	}

	//Removing Blank Data Key
	if len(dataResponse) > 0 {
		response.Data = dataResponse
	}

	c.JSON(http.StatusOK, response)
}

// ParseResponse from request sent internally
func ParseResponse(c *gin.Context, respBytes []byte, statusCode int) {

	apiCR := ValidateClientResponse(c, respBytes, statusCode)

	if apiCR != nil {
		GenerateResponse(c, apiCR.Response)
	}
}

// Exposed Method to Send a HTTP Request
func SendRequest(c *gin.Context, serviceType ServiceType, url string, requestType RequestType, headers map[string]interface{}, params map[string]string, body interface{}) {
	respBytes, statusCode, err := GetRequestResponse(serviceType, url, requestType, headers, params, body)
	if respBytes == nil {
		//If API fails or any other error
		GeneralAPIError(c, err.Error())
		return
	}

	//Parse response
	ParseResponse(c, respBytes, statusCode)

}

// Exposed Method to send a POST request to external services
func SendPostRequestToExternalService(url string, headers map[string]interface{}, body interface{}) ([]byte, int, error) {
	//Create internal API client
	client := NewAPIClient()

	options := PostRequestOptions{
		Url:           url,
		CustomHeaders: headers,
		Body:          body,
	}

	respBytes, statusCode, err := client.PostRequest(&options, BodyTypeRaw)

	if err != nil || respBytes == nil {
		return nil, DefaultStatusCode, err
	}

	return respBytes, statusCode, nil
}
