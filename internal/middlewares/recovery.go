package middlewares

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/utils"
	"github.com/gin-gonic/gin"
)

// CustomRecoveryMiddleware is a Gin middleware to handle panics
func CustomRecoveryMiddleware(c *gin.Context, err interface{}) {

	// Send Internal server error with error_message
	utils.GeneralAPIInternalError(c, utils.ErrorSomethingWentWrong)
}
