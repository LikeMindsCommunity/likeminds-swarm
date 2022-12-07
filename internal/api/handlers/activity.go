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

func createActivity(helper interfaces.ActivityHelper, action string, entity_id primitive.ObjectID, entity_type string,
	api_key string, action_by string, action_on string, cta_data map[string]interface{}) error {
	switch action {
	case constants.LikeAction, constants.TagAction:
		// activity filter data
		activity_filter_data := gin.H{
			"entity_id":   entity_id,
			"entity_type": entity_type,
			"action":      action,
			"api_key":     api_key,
			"action_by":   action_by,
			"action_on":   action_on,
		}

		// fetch activity using helper method
		activity_results, err := helper.FindActivityHelper(activity_filter_data)
		if err != nil {
			return err
		}

		// checking of existing activity
		if len(activity_results) == 0 {
			_, err := helper.CreateActivityHelper(action_by, []string{action_on}, api_key, entity_type, entity_id,
				action, cta_data)
			if err != nil {
				return err
			}
		}

	case constants.AlsoCommentAction:
		// activity filter data
		activity_filter_data := gin.H{
			"entity_id":   entity_id,
			"entity_type": entity_type,
			"action":      action,
			"api_key":     api_key,
		}

		// fetch activity using helper method
		activity_results, err := helper.FindActivityHelper(activity_filter_data)
		if err != nil {
			return err
		}

		// checking of existing activity
		if len(activity_results) == 0 {
			_, err := helper.CreateActivityHelper(action_by, []string{}, api_key, entity_type, entity_id,
				constants.AlsoCommentAction, cta_data)
			if err != nil {
				return err
			}
		} else {
			activity_data := activity_results[0]

			// activity update data
			activity_update_data := gin.H{
				"$push": gin.H{
					"action_on": activity_data.ActionBy,
				},
				"$set": gin.H{
					"action_by": action_by,
				},
			}

			// update activity using the helper method
			err = helper.UpdateActivityByIdHelper(activity_data.ID, activity_update_data)
			if err != nil {
				return err
			}
		}

	case constants.CommentAction, constants.DeleteAction:
		_, err := helper.CreateActivityHelper(action_by, []string{action_on}, api_key, entity_type, entity_id,
			action, cta_data)
		if err != nil {
			return err
		}
	}

	return nil
}

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
	_, err := handlers.activityHelper.CreateActivityHelper(headers[utils.HeadersMemberId], []string{user_id}, headers[utils.HeadersApiKey],
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
