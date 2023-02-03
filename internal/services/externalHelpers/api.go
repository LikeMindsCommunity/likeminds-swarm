package externalHelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type BodyType int

const (
	BodyTypeRaw BodyType = iota
	BodyTypeFormUrlEncoded
)

const DefaultStatusCode = -1

type APIClient struct {
	CaravanServiceBaseURL string
	HTTPClient            *http.Client
}

type GetRequestOptions struct {
	Url           string
	Params        map[string]string
	CustomHeaders map[string]interface{}
}

type PostRequestOptions struct {
	Url           string
	Params        map[string]string
	Body          interface{}
	CustomHeaders map[string]interface{}
}

// Exposed Method to Get Caravan Service URL
func GetCaravanServiceBaseUrl() string {
	CaravanServiceBaseURL := os.Getenv("CARAVAN_SERVICE_URL")

	if len(CaravanServiceBaseURL) == 0 {
		CaravanServiceBaseURL = "https://beta.likeminds.community"
	}

	return CaravanServiceBaseURL
}

// Exposed Method to Create New API Client
func NewAPIClient() *APIClient {
	return &APIClient{
		CaravanServiceBaseURL: GetCaravanServiceBaseUrl(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// Exposed Method to Add Headers to a Request
func AddHeaders(req *http.Request, headers map[string]interface{}) {
	for k, v := range headers {
		req.Header.Add(k, v.(string))
	}
}

// Exposed Method to Add Params to a Request
func AddParams(req *http.Request, params map[string]string) {
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
}

// Exposed Method to Update Post Body to a Request
func UpdateBody(pro *PostRequestOptions, body_type BodyType) (*http.Request, error) {

	var req *http.Request

	switch body_type {
	case BodyTypeRaw:

		data, err := json.Marshal(pro.Body)

		if err != nil {
			return nil, err
		}

		req, err = http.NewRequest(http.MethodPost, pro.Url, bytes.NewBuffer(data))

		if err != nil {
			return nil, err
		}

	case BodyTypeFormUrlEncoded:

		var err error

		body, _ := json.Marshal(pro.Body)
		payload := convertToFormURLEncoded(&body)

		req, err = http.NewRequest(http.MethodPost, pro.Url, strings.NewReader(payload.Encode()))

		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	}

	return req, nil
}

// Internal Method to convert raw body to url form encoded body
func convertToFormURLEncoded(body *[]byte) url.Values {
	// datamap | converts incoming request body into a map
	var datamap map[(string)]interface{}
	json.Unmarshal(*body, &datamap)

	// payload | form-url-encode paylaod
	payload := url.Values{}

	// loop over map and fill payload data
	for key, value := range datamap {
		payload.Set(key, fmt.Sprintf("%+v", value))
	}

	return payload
}

// Internal Method to send an http Request
func (c *APIClient) sendRequest(req *http.Request) ([]byte, int, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusInternalServerError {
		return nil, DefaultStatusCode, fmt.Errorf("unknown error, status code: %d", resp.StatusCode)
	}
	//Defer close error
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, DefaultStatusCode, err
	}
	return respBytes, resp.StatusCode, nil
}

// Exposed Method to send GET request
func (c *APIClient) GetRequest(gro *GetRequestOptions) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodGet, gro.Url, nil)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	params := gro.Params
	if params != nil {
		AddParams(req, params)
	}

	headers := gro.CustomHeaders
	if headers != nil {
		AddHeaders(req, headers)
	}

	respBytes, statusCode, err := c.sendRequest(req)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	return respBytes, statusCode, nil
}

// Exposed Method to send Post request
func (c *APIClient) PostRequest(pro *PostRequestOptions, body_type BodyType) ([]byte, int, error) {

	req, err := UpdateBody(pro, body_type)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	params := pro.Params
	if params != nil {
		AddParams(req, params)
	}

	headers := pro.CustomHeaders
	if headers != nil {
		AddHeaders(req, headers)
	}

	respBytes, statusCode, err := c.sendRequest(req)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	return respBytes, statusCode, nil
}

// Exposed Method to send Put request
func (c *APIClient) PutRequest(pro *PostRequestOptions) ([]byte, int, error) {
	jsonData, err := json.Marshal(pro.Body)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	req, err := http.NewRequest(http.MethodPut, pro.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	params := pro.Params
	if params != nil {
		AddParams(req, params)
	}

	headers := pro.CustomHeaders
	if headers != nil {
		AddHeaders(req, headers)
	}

	respBytes, statusCode, err := c.sendRequest(req)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	return respBytes, statusCode, nil
}

// Exposed Method to send Delete request
func (c *APIClient) DeleteRequest(pro *PostRequestOptions) ([]byte, int, error) {
	jsonData, err := json.Marshal(pro.Body)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	req, err := http.NewRequest(http.MethodDelete, pro.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	params := pro.Params
	if params != nil {
		AddParams(req, params)
	}

	headers := pro.CustomHeaders
	if headers != nil {
		AddHeaders(req, headers)
	}

	respBytes, statusCode, err := c.sendRequest(req)
	if err != nil {
		return nil, DefaultStatusCode, err
	}

	return respBytes, statusCode, nil
}
