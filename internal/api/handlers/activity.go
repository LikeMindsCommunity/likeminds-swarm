package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func fetchActivity(helper interfaces.ActivityHelper, activity_id string) (*entities.Activity, error) {
	// activity filter data
	activity_filter_data := gin.H{
		"_id": activity_id,
	}

	// fetch activity using helper method
	activity_results, err := helper.FindActivityHelper(activity_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of activity
	if len(activity_results) == 0 {
		return nil, fmt.Errorf("invalid activity_id sent")
	}

	return &activity_results[0], nil
}

func createActivity(handler FeedHandlers, action string, entity_id primitive.ObjectID, entity_type string,
	community_id int, action_by string, action_on string, cta_data map[string]interface{}) (interface{}, error) {
	var newActivityId interface{}

	switch action {
	case constants.LikeAction, constants.TagAction:
		// activity filter data
		activity_filter_data := gin.H{
			"entity_id":    entity_id,
			"entity_type":  entity_type,
			"action":       action,
			"community_id": community_id,
			"action_by":    action_by,
			"action_on":    action_on,
		}

		// fetch activity using helper method
		activity_results, err := handler.activityHelper.FindActivityHelper(activity_filter_data, gin.H{})
		if err != nil {
			return nil, err
		}

		// checking of existing activity
		if len(activity_results) == 0 {
			activityId, err := handler.activityHelper.CreateActivityHelper(action_by, []string{action_on}, community_id,
				entity_type, entity_id, action, cta_data)
			if err != nil {
				return nil, err
			}

			newActivityId = activityId
		} else {
			newActivityId = activity_results[0].ID
		}

	case constants.AlsoCommentAction:
		// activity filter data
		activity_filter_data := gin.H{
			"entity_id":    entity_id,
			"entity_type":  entity_type,
			"action":       action,
			"community_id": community_id,
		}

		// fetch activity using helper method
		activity_results, err := handler.activityHelper.FindActivityHelper(activity_filter_data, gin.H{})
		if err != nil {
			return nil, err
		}

		// checking of existing activity
		if len(activity_results) == 0 {
			activityId, err := handler.activityHelper.CreateActivityHelper(action_by, []string{}, community_id, entity_type,
				entity_id, constants.AlsoCommentAction, cta_data)
			if err != nil {
				return nil, err
			}

			newActivityId = activityId
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
			err = handler.activityHelper.UpdateActivityByIdHelper(activity_data.ID, activity_update_data)
			if err != nil {
				return nil, err
			}

			newActivityId = activity_data.ID
		}

	case constants.CommentAction, constants.DeleteAction, constants.CreatePostPermitAddedAction, constants.CreatePostPermitRemovedAction,
		constants.CreateCommentPermissionAddedAction, constants.CreateCommentPermitRemovedAction:
		activityId, err := handler.activityHelper.CreateActivityHelper(action_by, []string{action_on}, community_id,
			entity_type, entity_id, action, cta_data)
		if err != nil {
			return nil, err
		}

		newActivityId = activityId
	}

	// send notification
	SendNotification(newActivityId.(primitive.ObjectID), handler)

	return newActivityId, nil
}

func (handlers *FeedHandlers) ExternalCreateActivity(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	user_id := c.Param("user_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

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

	if user_id == "" {
		utils.GeneralAPIValidationError(c, "Send valid user_id")
		return
	}

	// create activity using the helper method
	_, err := createActivity(*handlers, externalCreateActivityRequest.Action, primitive.NilObjectID,
		constants.UserEntityType, community_id, headers[utils.HeadersMemberId], user_id, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
