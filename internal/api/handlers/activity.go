package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (handlers *activityHandlers) ExternalCreateActivity(c *gin.Context) {
	// validation of request body
	var externalCreateActivityRequest requests.CreateActivityRequest
	if err := c.ShouldBindJSON(&externalCreateActivityRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of valid actions
	var validActions = []string{constants.CreatePostPermitAddedAction, constants.CreatePostPermitRemovedAction,
		constants.CreateCommentPermissionAddedAction, constants.CreateCommentPermitRemovedAction}
	var isActionValid bool = false

	for _, value := range validActions {
		if value == externalCreateActivityRequest.Action {
			isActionValid = true
		}
	}

	if !isActionValid {
		utils.GeneralAPIValidationError(c, "Invalid action sent")
		return
	}

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	user_id := c.Param("user_id")

	if user_id == "" {
		utils.GeneralAPIValidationError(c, "Send valid user_id")
		return
	}

	// create activity using the helper method
	_, err := handlers.activityHelper.CreateActivityHelper(headers[utils.HeadersMemberId], user_id, headers[utils.HeadersApiKey],
		constants.UserEntityType, primitive.NilObjectID, externalCreateActivityRequest.Action, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

type activityHandlers struct {
	activityHelper interfaces.ActivityHelper
}

func NewActivityHandlers(activityHelper interfaces.ActivityHelper) *activityHandlers {
	return &activityHandlers{
		activityHelper: activityHelper,
	}
}
