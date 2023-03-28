package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error Messages
const (
	NotAuthorizedError = "You are not authorized to perform this operation."
)

// Exposed Method to send General Validation Error in API Response
func GeneralAPIValidationError(c *gin.Context, errorMessage string) {
	GeneralAPIError(c, errorMessage, http.StatusBadRequest)
}

// Exposed Method to send General Internal Error in API Response
func GeneralAPIInternalError(c *gin.Context, errorMessage string) {
	GeneralAPIError(c, errorMessage, http.StatusInternalServerError)
}

// Exposed Method to send General Validation Error for Middlewares
func MiddlewareGeneralValidationError(c *gin.Context, errorMessage string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"success":       false,
		"error_message": errorMessage,
	})
}

// Exposed Method to send General Error in API Response
func GeneralAPIError(c *gin.Context, errorMessage string, statusCode int) {
	c.JSON(statusCode, gin.H{
		"success":       false,
		"error_message": errorMessage,
	})
}
