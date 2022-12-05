package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GeneralAPIValidationError(c *gin.Context, errorMessage string) {
	GeneralAPIError(c, errorMessage, http.StatusBadRequest)
}

func GeneralAPIInternalError(c *gin.Context, errorMessage string) {
	GeneralAPIError(c, errorMessage, http.StatusInternalServerError)
}

func GeneralAPIError(c *gin.Context, errorMessage string, statusCode int) {
	c.JSON(statusCode, gin.H{
		"success":       false,
		"error_message": errorMessage,
	})
}
