package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// CustomRecoveryMiddleware is a Gin middleware to handle panics
func CustomRecoveryMiddleware(c *gin.Context, err interface{}) {

	// Send Internal server error with error_message
	utils.GeneralAPIInternalError(c, utils.ErrorSomethingWentWrong)
}
