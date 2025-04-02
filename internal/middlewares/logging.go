package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// responseBodyWriter | Custom Response Writer
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write | Custom Write Method for responseBodyWriter
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Method to process API request to log
func processRequest(c *gin.Context) interface{} {
	requestBodyData := gin.H{}

	// Reading request body
	requestBody, err := io.ReadAll(c.Request.Body)

	// Updating request body after read
	c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

	// Unmarshalling request body
	if err == nil {
		_ = json.Unmarshal(requestBody, &requestBodyData)
	}

	return gin.H{
		"host":         c.Request.Host,
		"absolute_uri": c.Request.RequestURI,
		"method":       c.Request.Method,
		"headers":      c.Request.Header,
		"body":         requestBodyData,
	}
}

// sanitizeRequestHeaders removes sensitive headers and returns a sanitized copy
func sanitizeRequestHeaders(requestData gin.H) {
	if headers, ok := requestData["headers"].(http.Header); ok {
		headersCopy := make(map[string]string)
		for key, values := range headers {
			if len(values) > 0 {
				// Remove hyphens from header keys
				sanitizedKey := strings.ReplaceAll(key, "-", "_")
				headersCopy[sanitizedKey] = values[0]
			}
		}
		requestData["headers"] = headersCopy
	}
}
// logToAppInsights marshals the data to JSON and logs it to Azure Application Insights
func logToAppInsights(data gin.H) {
	client := logging.GetAppInsightsClient()

	request := appinsights.NewRequestTelemetry(
		fmt.Sprint(data["request"].(gin.H)["method"]),
		fmt.Sprint(data["request"].(gin.H)["absolute_uri"]),
		time.Duration(0), // Duration will be set later
		fmt.Sprint(data["response"].(gin.H)["http_response_code"]),
	)

	if meta, ok := data["meta"].(gin.H); ok {
		if latency, ok := meta["latency"].(time.Duration); ok {
			request.Duration = latency
			if clientIP, ok := meta["client_ip"].(string); ok {
				request.Source = clientIP
			}
		}
	}

	// Add the IST timestamp to request.Properties
	istLocation, _ := time.LoadLocation("Asia/Kolkata")
	currentTimeIST := time.Now().In(istLocation)
	formattedTimeIST := currentTimeIST.Format("2006-01-02 15:04:05")
	request.Properties["timestamp_IST"] = formattedTimeIST

	// Add all fields from data to request.Properties
	for key, value := range data {
		switch v := value.(type) {
		case gin.H:
			// Serialize nested maps to JSON
			nestedJSON, _ := json.Marshal(v)
			request.Properties[key] = string(nestedJSON)
		default:
			// Add other types directly
			request.Properties[key] = fmt.Sprint(v)
		}
	}

	// Set success based on status code
	if statusCode, ok := data["response"].(gin.H)["http_response_code"].(int); ok {
		request.Success = statusCode >= 200 && statusCode < 400
	}

	client.Track(request)
}

// LoggingMiddleware will log the request and response of API
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/" {

			c.Next()

		} else {

			data := gin.H{}

			// Starting time
			startTime := time.Now()

			// Implementing custom response body writer in the context
			w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
			c.Writer = w

			// Updating Request Data
			data["request"] = processRequest(c)

			sanitizeRequestHeaders(data["request"].(gin.H))

			// Processing request
			c.Next()

			// End Time
			endTime := time.Now()

			response := gin.H{}
			statusCode := c.Writer.Status()

			// Unmarshalling Request Response
			_ = json.Unmarshal(w.body.Bytes(), &response)

			// Updating Request Response
			data["response"] = gin.H{
				"http_response_code": statusCode,
				"content":            response,
			}

			if statusCode < http.StatusBadRequest {
				data["response"].(gin.H)["content"] = gin.H{}
			}

			// Updating Request Meta Data
			data["meta"] = gin.H{
				"latency":   endTime.Sub(startTime),
				"client_ip": c.ClientIP(),
			}

			// Marshalling the final Data
			if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
				// Logging the generated request data as Info
				logging.InfoWithFields(data)
				logToAppInsights(data)
			} else {
				// Logging the generated request data as Error
				logging.ErrorWithFields(data)
				logToAppInsights(data)
			}

			c.Next()
		}
	}
}
