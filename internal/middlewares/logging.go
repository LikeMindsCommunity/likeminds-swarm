package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
			marshelledData, _ := json.Marshal(data)

			if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
				// Logging the generated request data as Info
				logging.Info(string(marshelledData))
			} else {
				// Logging the generated request data as Error
				logging.Error(string(marshelledData))
			}

			c.Next()
		}
	}
}
