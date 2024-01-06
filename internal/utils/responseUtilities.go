package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error Messages
const (
	NotAuthorizedError      = "You are not authorized to perform this operation."
	InvalidRequestError     = "Invalid request."
	InvalidPostIDError      = "Invalid post_id sent."
	InvalidCommentIDError   = "Invalid comment_id sent."
	NsfwContentInImageError = "This post could not be submitted as %simage/s seems to contain NSFW content."
)

// Exposed Method to send General Validation Error in API Response
func GeneralAPIValidationError(c *gin.Context, errorMessage string) {
	GeneralAPIError(c, errorMessage, http.StatusBadRequest)
}

// Exposed Method to send General Internal Error in API Response
func GeneralAPIInternalError(c *gin.Context, errorMessage string) {
	GeneralAPIError(c, errorMessage, http.StatusInternalServerError)
}

// Exposed Method to send General Error in API Response
func GeneralAPIError(c *gin.Context, errorMessage string, statusCode int) {
	c.JSON(statusCode, gin.H{
		"success":       false,
		"error_message": errorMessage,
	})
}

// Exposed method to send custom error in API Response
func CustomAPIErrorWithMeta(c *gin.Context, statusCode int, errorMessage string, errorMeta gin.H) {

	errorResponse := gin.H{
		"success":       false,
		"error_message": errorMessage,
		"error_meta":    errorMeta,
	}

	c.JSON(statusCode, errorResponse)
}

// Exposed Method to send successfull API Response
func GenereateSuccessResponse(c *gin.Context, dataResponse gin.H) {

	if dataResponse == nil {
		dataResponse = gin.H{}
	}

	// set success flag
	dataResponse["success"] = true

	// return final response
	c.JSON(http.StatusOK, dataResponse)
}
