package handlers

import (
	"fmt"
	"net/http"

	"github.com/aquilax/truncate"
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse User activity list
func parseUserActivity(handler FeedHandlers, activities []entities.Activity) ([]requests.UserActivityResponse, error) {
	response := []requests.UserActivityResponse{}

	for _, activity := range activities {
		activityUserData := getActivityUserData(activity)

		activityEntityData, err := getEntityData(handler, activity.EntityType, activity.EntityID)
		if err != nil {
			return response, err
		}

		activityText, err := getActivityText(activityUserData, activityEntityData, activity)
		if err != nil {
			return response, err
		}

		activity = activity
		response = append(response, requests.UserActivityResponse{
			ID:                 activity.ID,
			ActionBy:           activity.ActionBy,
			ActionOn:           activity.ActionOn,
			EntityType:         activity.EntityType,
			EntityID:           activity.EntityID,
			EntityOwnerID:      activity.EntityOwnerID,
			Action:             activity.Action,
			CTA:                activity.CTA,
			IsRead:             activity.IsRead,
			CreatedAt:          int(activity.CreatedAt.UnixMilli()),
			UpdatedAt:          int(activity.UpdatedAt.UnixMilli()),
			ActivityUserData:   activityUserData,
			ActivityEntityData: activityEntityData,
			ActivityText:       activityText,
		})
	}
	return response, nil
}

func getActivityUserData(activity entities.Activity) map[string]interface{} {
	activityUserUID := activity.ActionBy[len(activity.ActionBy)-1]
	activityUserData := map[string]interface{}{}
	isSuccess := false

	isSuccess, activityUserData[activityUserUID] = externalHelpers.FetchMemberMeta([]string{activityUserUID}, activity.ActionOn, activity.CommunityID)
	if isSuccess {
		return activityUserData
	}

	return nil
}

func getEntityData(handler FeedHandlers, entityType constants.EntityType, entityID primitive.ObjectID) (interface{}, error) {

	switch entityType {
	case constants.Post:

		postFilter := gin.H{
			"_id": entityID,
		}

		postData, err := handler.postHelper.FindPostHelper(postFilter, gin.H{})
		if err != nil {
			return nil, err
		}

		return postData[0], nil

	case constants.Comment:

		commentFilter := gin.H{
			"_id": entityID,
		}

		commentData, err := handler.commentHelper.FindCommentHelper(commentFilter, gin.H{})
		if err != nil {
			return nil, err
		}

		return commentData[0], nil
	}

	return nil, nil
}

func getActivityText(activityUserData map[string]interface{}, activityEntityData interface{}, activity entities.Activity) (string, error) {
	activityText := ""

	switch activity.Action {
	case constants.CreatePostPermitAdded:
		activityText += "You now have the permission to create posts in the community. Start posting now."
		return activityText, nil

	case constants.CreatePostPermitRemoved:
		activityText += "Your permission to create posts in the community has been removed."
		return activityText, nil

	case constants.CreateCommentPermitAdded:
		activityText += "You now have the permission to add comments on the posts. Start engaging now."
		return activityText, nil

	case constants.CreateCommentPermitRemoved:
		activityText += "Your permission to add comments and replies to the posts has been removed."
		return activityText, nil

	case constants.CMDeletedPost:
		activityText += "Your post has been deleted as it violates community guidelines. Reason: "
		activityText += activityEntityData.(entities.Post).DeleteReason
		return activityText, nil

	case constants.CMDeletedComment:
		activityText += "Your Reply has been deleted as it violates community guidelines. Reason: "
		activityText += activityEntityData.(entities.Comment).DeleteReason
		return activityText, nil

	case constants.LikeOnPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityByUserDataName := activityByUserData.(*externalHelpers.MemberMetaResponse).Members[0].Name

		activityText += activityByUserDataName

		// add condition for and n others

		activityText += " liked your post \""

		postDataText := activityEntityData.(entities.Post).Text
		postDataTextTrimmed := truncate.Truncate(postDataText, 60, "...", truncate.PositionEnd)
		activityText += postDataTextTrimmed + "\""

		return activityText, nil

	case constants.CommentOnPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityByUserDataName := activityByUserData.(*externalHelpers.MemberMetaResponse).Members[0].Name

		activityText += activityByUserDataName
		// add condition for and n others
		activityText += " commented on your post \""
		postDataText := activityEntityData.(entities.Post).Text
		postDataTextTrimmed := truncate.Truncate(postDataText, 60, "...", truncate.PositionEnd)
		activityText += postDataTextTrimmed + "\""

		return activityText, nil

	case constants.LikeOnComment:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityByUserDataName := activityByUserData.(*externalHelpers.MemberMetaResponse).Members[0].Name

		activityText += activityByUserDataName

		// add condition for and n others

		activityText += " liked on your comment \""

		commentDataText := activityEntityData.(entities.Comment).Text
		commentDataTextTrimmed := truncate.Truncate(commentDataText, 60, "...", truncate.PositionEnd)
		activityText += commentDataTextTrimmed + "\""

		return activityText, nil

	case constants.CommentOnComment:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityByUserDataName := activityByUserData.(*externalHelpers.MemberMetaResponse).Members[0].Name

		activityText += activityByUserDataName

		// add condition for and n others

		activityText += " replied on your comment \""

		commentDataText := activityEntityData.(entities.Comment).Text
		commentDataTextTrimmed := truncate.Truncate(commentDataText, 60, "...", truncate.PositionEnd)
		activityText += commentDataTextTrimmed + "\""

		return activityText, nil

	case constants.TaggedInPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityByUserDataName := activityByUserData.(*externalHelpers.MemberMetaResponse).Members[0].Name

		activityText += activityByUserDataName
		activityText += " tagged you in their post \""

		postDataText := activityEntityData.(entities.Post).Text
		postDataTextTrimmed := truncate.Truncate(postDataText, 60, "...", truncate.PositionEnd)
		activityText += postDataTextTrimmed + "\""

		return activityText, nil

	case constants.TaggedInPostComment:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityByUserDataName := activityByUserData.(*externalHelpers.MemberMetaResponse).Members[0].Name

		activityText += activityByUserDataName
		activityText += " tagged you in their comment \""

		commentDataText := activityEntityData.(entities.Comment).Text
		commentDataTextTrimmed := truncate.Truncate(commentDataText, 60, "...", truncate.PositionEnd)
		activityText += commentDataTextTrimmed + "\""

		return activityText, nil
	}

	return activityText, nil
}

// Internal Method to fetch activity using activity_id
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

// Exposed Helper Method to Create Activity
func (handlers *FeedHandlers) CreateActivity(communityID int, actionBy []string, actionOn string, entityType constants.EntityType,
	entityID primitive.ObjectID, entityOwnerID string, action constants.ActivityAction, ctaData map[string]interface{}, isRead bool) (interface{}, error) {
	cta := fetchActivityCtaForAction(action, ctaData)

	switch action {
	case constants.CreatePostPermitAdded,
		constants.CreatePostPermitRemoved,
		constants.CreateCommentPermitAdded,
		constants.CreateCommentPermitRemoved,
		constants.CMDeletedPost,
		constants.CMDeletedComment,
		constants.LikeOnPost,
		constants.CommentOnPost,
		constants.LikeOnComment,
		constants.CommentOnComment,
		constants.TaggedInPost,
		constants.TaggedInPostComment:

		activityID, err := handlers.activityHelper.CreateActivityHelper(communityID, actionBy, actionOn, entityType, entityID, entityOwnerID, action, cta, isRead)
		return activityID, err

	}

	return nil, nil
}

// FetchActivityCtaForAction | get CTA corresponding to action
func fetchActivityCtaForAction(action constants.ActivityAction, ctaData map[string]interface{}) string {
	var cta string = ""

	switch action {
	case
		constants.LikeOnPost,
		constants.CommentOnPost,
		constants.LikeOnComment,
		constants.CommentOnComment,
		constants.TaggedInPost,
		constants.TaggedInPostComment,
		constants.AlsoCommentOnPost:
		cta = parseCTAData(ctaData)

	case constants.CreatePostPermitAdded:
		cta = utils.CreatePostRoute

	case constants.CreateCommentPermitAdded:
		cta = utils.HomeFeedRoute
	}

	return cta
}

// Internal Method to parse CTA Data
func parseCTAData(cta_data map[string]interface{}) string {
	var cta string = ""

	if entity_type, ok := cta_data["entity_type"]; ok {
		if post_id, ok := cta_data["post_id"]; ok {
			switch entity_type {
			case constants.PostEntityType:
				cta = fmt.Sprintf(utils.PostDetailRoute, post_id)

			case constants.CommentEntityType:
				if comment_id, ok := cta_data["comment_id"]; ok {
					cta = fmt.Sprintf(utils.CommentDetailRoute, post_id, comment_id)
				}
			}
		}
	}

	return cta
}

// // Exposed Method to create new Activity
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
	var isActionValid bool = false
	var action constants.ActivityAction = constants.DefaultAction

	switch externalCreateActivityRequest.Action {
	case constants.CreatePostPermitAddedAction:
		action = constants.CreatePostPermitAdded
		isActionValid = true
	case constants.CreatePostPermitRemovedAction:
		action = constants.CreatePostPermitRemoved
		isActionValid = true
	case constants.CreateCommentPermissionAddedAction:
		action = constants.CreateCommentPermitAdded
		isActionValid = true
	case constants.CreateCommentPermitRemovedAction:
		action = constants.CreateCommentPermitRemoved
		isActionValid = true
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
	_, err := handlers.CreateActivity(community_id, []string{headers[utils.HeadersMemberId]}, user_id, constants.User, primitive.NilObjectID, user_id, action, gin.H{}, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// 	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// FetchUserActivity | method to Fetch User Activity
func (handlers *FeedHandlers) FetchUserActivity(c *gin.Context) {

	// fetch url params and headers
	headers := utils.GetHeaders(c)
	userID := c.Param("user_id")

	if userID != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// validation of api_key
	communityID := externalHelpers.GetCommunityId(c)
	if communityID == externalHelpers.DefaultCommunityId {
		return
	}

	// activity filter data
	activityFilterData := gin.H{
		"action_on":    userID,
		"community_id": communityID,
	}

	// filter options
	activityFilterOptions, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch activity using helper method
	activityResults, err := handlers.activityHelper.FindActivityHelper(activityFilterData, activityFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse user activity response
	activityResponse, err := parseUserActivity(*handlers, activityResults)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// 	// return final response
	c.JSON(http.StatusOK, gin.H{"success": true, "activities": activityResponse})
}

// UserActivityMarkRead | Mark user activity as read
func (handlers *FeedHandlers) UserActivityMarkRead(c *gin.Context) {

	headers := utils.GetHeaders(c)
	userID := c.Param("user_id")
	activityID := c.Param("activity_id")

	// validation of api_key
	communityID := externalHelpers.GetCommunityId(c)
	if communityID == externalHelpers.DefaultCommunityId {
		return
	}

	if userID != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	activityFilterData := gin.H{
		"_id":       activityID,
		"action_on": userID,
	}

	// fetch activity to check activity owner - action_on
	activity, err := handlers.activityHelper.FindActivityHelper(activityFilterData, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activity == nil {
		utils.GeneralAPIValidationError(c, "Activity not found or You are not authorized to perform this operation.")
		return
	}

	// activity update data
	activityUpdateData := gin.H{
		"$set": gin.H{
			"is_read": true,
		},
	}

	// update comment data
	err = handlers.activityHelper.UpdateActivityByIDHelper(activity[0].ID, activityUpdateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UserActivityFeedUnreadCount | Get user activity feed unread count
func (handlers *FeedHandlers) UserActivityFeedUnreadCount(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userID := c.Param("user_id")

	// validation of api_key
	communityID := externalHelpers.GetCommunityId(c)
	if communityID == externalHelpers.DefaultCommunityId {
		return
	}

	if userID != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// activity filter data
	activityFilterData := gin.H{
		"community_id": communityID,
		"action_on":    userID,
		"is_read":      false,
	}

	// fetch activity using helper method
	activity_unread_count, err := handlers.activityHelper.CountActivityHelper(activityFilterData)
	if err != nil {
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{"success": true, "count": activity_unread_count})
}
