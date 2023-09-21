package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse User activity list
func parseUserActivity(handler FeedHandlers, activities []entities.Activity,
	apiRevampV1Check bool, uuid string) ([]interface{}, interface{}, interface{}, interface{}, error) {

	response := []interface{}{}
	userDatas := make(map[string]interface{})
	topicDatas := map[string]interface{}{}
	widgetDatas := map[string]interface{}{}

	for _, activity := range activities {
		activityUserData, activityUserUID := getActivityUserData(activity)

		activityEntityData, err := getEntityData(handler, activity.EntityType, activity.EntityID, activity.CommunityID,
			apiRevampV1Check)
		if err != nil {
			return response, userDatas, topicDatas, widgetDatas, err
		}

		activityText, err := getActivityText(activityUserData, activityEntityData, activity)
		if err != nil {
			return response, userDatas, topicDatas, widgetDatas, err
		}

		userActivity := requests.UserActivityResponse{
			ID:                 activity.ID,
			ActionBy:           activity.ActionBy,
			ActionOn:           activity.ActionOn,
			EntityID:           activity.EntityID,
			EntityOwnerID:      activity.EntityOwnerID,
			UUID:               activity.EntityOwnerID,
			CTA:                activity.CTA,
			IsRead:             activity.IsRead,
			CreatedAt:          int(activity.CreatedAt.UnixMilli()),
			UpdatedAt:          int(activity.UpdatedAt.UnixMilli()),
			ActivityEntityData: activityEntityData,
			ActivityText:       activityText,
		}

		if apiRevampV1Check {

			// API Revamp V1 Response
			response = append(response, requests.UserActivityResponseV1{
				UserActivityResponse: userActivity,
				EntityType:           enums.NewEntityTypeFromInt(int(activity.EntityType)),
				Action:               enums.NewActivityActionFromInt(int(activity.Action)),
			})

		} else {

			// Old User Activity Response
			response = append(response, requests.UserActivityResponseOld{
				UserActivityResponse: userActivity,
				EntityType:           activity.EntityType,
				Action:               activity.Action,
			})
		}

		userDatas[activityUserUID] = activityUserData[activityUserUID]

		if activity.EntityType == constants.Post {
			topicsData, _ := parseTopicsResponse(handler.topicHelper, activityEntityData.(requests.PostResponse).Topics,
				activity.CommunityID)

			widgetIds := getWidgetIdsFromAttachments(activityEntityData.(requests.PostResponse).Attachments)
			widgetsData, _ := parseWidgetsResponse(&handler, widgetIds, activity.CommunityID, uuid)

			for topicId, topicData := range topicsData {
				topicDatas[topicId] = topicData
			}

			for widgetId, widgetData := range widgetsData {
				widgetDatas[widgetId] = widgetData
			}
		}
	}

	return response, userDatas, topicDatas, widgetDatas, nil
}

func getActivityUserData(activity entities.Activity) (map[string]interface{}, string) {
	activityUserUID := activity.ActionBy[len(activity.ActionBy)-1]
	activityUserData := map[string]interface{}{}
	isSuccess := false

	isSuccess, activityUserData[activityUserUID] = externalHelpers.FetchMemberMeta([]string{activityUserUID}, activity.ActionOn, activity.CommunityID)
	activityUserData[activityUserUID] = activityUserData[activityUserUID].(*externalHelpers.MemberMetaResponse).Members[0]
	if isSuccess {
		return activityUserData, activityUserUID
	}

	return nil, ""
}

func getEntityData(handler FeedHandlers, entityType constants.EntityType, entityID primitive.ObjectID, communityID int,
	apiRevampV1Check bool) (interface{}, error) {

	switch entityType {
	case constants.Post:
		postData, err := fetchMultiplePostsData(&handler, []string{entityID.Hex()}, communityID, "", false, "", "",
			apiRevampV1Check)
		if err != nil {
			return nil, err
		}

		return postData[entityID.Hex()], nil

	case constants.Comment:
		commentData, err := fetchMultipleCommentsData(&handler, []string{entityID.Hex()}, communityID, "", false, "", "",
			apiRevampV1Check)
		if err != nil {
			return nil, err
		}

		return commentData[entityID.Hex()].CommentResponse, nil
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
		activityText += activityEntityData.(requests.PostResponse).DeleteReason + "."
		return activityText, nil

	case constants.CMDeletedComment:
		activityText += "Your reply has been deleted as it violates community guidelines. Reason: "
		activityText += activityEntityData.(requests.CommentResponse).DeleteReason + "."
		return activityText, nil

	case constants.LikeOnPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " liked your"

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil

	case constants.CommentOnPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " commented on your"

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil

	case constants.LikeOnComment:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " liked your comment"

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil

	case constants.CommentOnComment:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " replied on your comment"

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil

	case constants.TaggedInPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)

		activityText += " tagged you in their"

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil

	case constants.TaggedInPostComment:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)

		activityText += " tagged you in their comment"

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil

	case constants.AlsoCommentOnPost:
		activityByUserData := activityUserData[activity.ActionBy[len(activity.ActionBy)-1]]
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " also commented on"

		activityEntityOwnerUserData, activityEntityOwnerUserID := fetchActivityEntityOwnerUserData(activity)
		if activityEntityOwnerUserID != "" {
			activityText += " " + getUserRoute(activityEntityOwnerUserData) + "'s"
		}

		activityText += getEntityText(activity.EntityType, activityEntityData)

		return activityText, nil
	}

	return activityText, nil
}

func getUserRoute(activityByUserData interface{}) string {
	activityByUserDataEntity := activityByUserData.(externalHelpers.MemberMeta)
	userRouteString := "<<%s|route://user_profile/%s>>"

	return fmt.Sprintf(userRouteString, activityByUserDataEntity.Name, activityByUserDataEntity.UserUniqueId)
}

func getMultipleUserActivityText(activity entities.Activity) string {
	if len(activity.ActionBy) <= 1 {
		return ""
	}

	stringOneOther := " and 1 other"

	activityMembersTotalBarOne := len(activity.ActionBy) - 1

	if activityMembersTotalBarOne == 1 {
		return stringOneOther
	}

	nOtherActivityTemplate := " and %s others"
	nOtherActivityText := fmt.Sprintf(nOtherActivityTemplate, strconv.Itoa(activityMembersTotalBarOne))

	return nOtherActivityText
}

func fetchActivityEntityOwnerUserData(activity entities.Activity) (map[string]interface{}, string) {
	userData := map[string]interface{}{}
	userID := activity.EntityOwnerID
	isSuccess := false

	isSuccess, userData[userID] = externalHelpers.FetchMemberMeta([]string{userID}, activity.ActionOn, activity.CommunityID)
	userData[userID] = userData[userID].(*externalHelpers.MemberMetaResponse).Members[0]
	if isSuccess {
		return userData, userID
	}

	return nil, ""
}

func getEntityText(entityType constants.EntityType, activityEntityData interface{}) string {
	entityTextData := ""

	switch entityType {
	case constants.Post:
		entityTextData = activityEntityData.(requests.PostResponse).Text

	case constants.Comment:
		entityTextData = activityEntityData.(requests.CommentResponse).Text
	}

	// if post text is nil, add attachment type as text
	if entityType == constants.Post && entityTextData == "" {
		return " " + getPostAttachmentType(activityEntityData.(requests.PostResponse)) + "."
	}

	if entityType == constants.Post && entityTextData != "" {
		return " post \"" + entityTextData + "\""
	}

	if entityTextData == "" {
		return entityTextData + "."
	}

	activityText := " \"" + entityTextData + "\""

	return activityText
}

func getPostAttachmentType(postResponse requests.PostResponse) string {
	if len(postResponse.Attachments) == 0 {
		return ""
	}

	intAttachmentType := postResponse.Attachments[0].Type.ToInt()
	enumAttachmentType := enums.NewAttachmentTypeFromInt(intAttachmentType)

	return enumAttachmentType.ToString()
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

// CreateActivity | method to create an activity record
func (handlers *FeedHandlers) CreateActivity(communityID int, actionBy []string, actionOn string, entityType constants.EntityType,
	entityID primitive.ObjectID, entityOwnerID string, action constants.ActivityAction, ctaData map[string]interface{}, isRead bool, isDeleted bool) (interface{}, error) {

	if len(actionBy) > 0 && actionBy[0] == actionOn {
		return nil, nil
	}

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
		constants.TaggedInPostComment,
		constants.AlsoCommentOnPost:

		activityID, err := handlers.activityHelper.CreateActivityHelper(communityID, actionBy, actionOn, entityType, entityID, entityOwnerID, action, cta, isRead, isDeleted)

		handlers.pushActivitytoCache(activityID)

		return activityID, err

	}

	return nil, nil
}

func (handlers *FeedHandlers) CreateAlsoCommentedActivity(activityID interface{}, postData *entities.Post) {
	postCommentActivity, err := fetchActivity(handlers.activityHelper, activityID.(string))
	if err != nil {
		return
	}

	latestCommentUser := postCommentActivity.ActionBy[len(postCommentActivity.ActionBy)-1]
	previousCommentUsers := utils.RemoveAllOccurenceStringList(postCommentActivity.ActionBy, latestCommentUser)

	for _, previousCommentUser := range previousCommentUsers {
		_, err := handlers.CreateActivity(postData.ChatroomId, []string{latestCommentUser}, previousCommentUser, constants.Post, postData.ID, postData.UserId, constants.AlsoCommentOnPost, gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     postData.ID,
		}, false, false)

		if err != nil {
		}
	}
}

// DeleteActivity | delete activity records with filter
func (handlers *FeedHandlers) DeleteActivity(filter map[string]interface{}) {
	handlers.activityHelper.DeleteActivityHelper(filter)
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
	activityID, err := handlers.CreateActivity(community_id, []string{headers[utils.HeadersMemberId]}, user_id, constants.User, primitive.NilObjectID, user_id, action, gin.H{}, false, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activityID != nil {
		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
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

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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
		"is_deleted":   false,
	}

	activitySortKey := "updated_at"

	// filter options
	activityFilterOptions, err := generatePageFilterOptions(c, activitySortKey, OrderTypeDescending)
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
	activityResponse, userDatas, topicDatas, widgetDatas, err := parseUserActivity(*handlers, activityResults,
		apiRevampV1Check, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// 	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"activities": activityResponse,
		"users":      userDatas,
		"topics":     topicDatas,
		"widgets":    widgetDatas,
	})
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
	err = handlers.activityHelper.UpdateActivityByIDHelper(activity[0].ID, activityUpdateData, true)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	handlers.updateActivityInCache(activity[0].ID.Hex())

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
	activityUnreadCount, err := handlers.activityHelper.CountActivityHelper(activityFilterData)
	if err != nil {
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{"success": true, "count": activityUnreadCount})
}

// updateActivityInCache | update activity in cache storage
func (handlers *FeedHandlers) updateActivityInCache(activityID string) {
	cacheActivityKey := fmt.Sprintf("activity_{}", activityID)
	cacheActivityString := handlers.cacheHelper.Get(cacheActivityKey)

	if cacheActivityString == nil {
		return
	}
	activityFilter := gin.H{
		"_id": activityID,
	}
	activity, err := handlers.activityHelper.FindActivityHelper(activityFilter, gin.H{})
	if err != nil {
		return
	}
	activtyBytes, err := json.Marshal(activity[0])
	if err != nil {
		return
	}
	activityString := string(activtyBytes)

	handlers.cacheHelper.Set(cacheActivityKey, activityString, 0)
}

func (handlers *FeedHandlers) pushActivitytoCache(activityID interface{}) {
	activityFilter := gin.H{
		"_id": activityID,
	}
	activity, err := handlers.activityHelper.FindActivityHelper(activityFilter, gin.H{})
	if err != nil {
		return
	}

	userID := activity[0].ActionOn
	activtyBytes, err := json.Marshal(activity[0])
	if err != nil {
		return
	}
	activityString := string(activtyBytes)

	cacheUserActivityFeedKey := fmt.Sprintf("user_{}_activity_feed", userID)
	handlers.cacheHelper.LPush(cacheUserActivityFeedKey, activityID.(string), 20)

	cacheActivityKey := fmt.Sprintf("activity_{}", activityID.(string))
	handlers.cacheHelper.Set(cacheActivityKey, activityString, 0)
}

// WarmupUserActivityFeedCache | push user activity feed first page to cache
func (handlers *FeedHandlers) WarmupUserActivityFeedCache(communityID int, userID string) []entities.Activity {
	handlers.deleteUserActivityFeedCacheData(userID)

	userActivities := []entities.Activity{}

	// activity filter data
	activityFilterData := gin.H{
		"action_on":    userID,
		"community_id": communityID,
		"is_deleted":   false,
	}

	activityFilterOptions := gin.H{
		"": "",
	}

	// fetch activity using helper method
	activityResults, err := handlers.activityHelper.FindActivityHelper(activityFilterData, activityFilterOptions)
	if err != nil {
		return userActivities
	}

	handlers.createUserActivityFeedCacheData(userID, activityResults)

	return userActivities
}

func (handlers *FeedHandlers) deleteUserActivityFeedCacheData(userID string) {
	userActivityFeedKey := fmt.Sprintf("user_{}_activity_feed", userID)

	cacheUserActivityIDsString := handlers.cacheHelper.Get(userActivityFeedKey)
	cacheUserActivityIDs := [](string){cacheUserActivityIDsString.Val()}

	cacheActivityKeys := [](string){}
	for _, cacheUserActivityID := range cacheUserActivityIDs {
		cacheActivityKey := fmt.Sprintf("activity_{}", cacheUserActivityID)
		cacheActivityKeys = append(cacheActivityKeys, cacheActivityKey)
	}

	handlers.cacheHelper.DelMultiple(cacheActivityKeys)
	handlers.cacheHelper.Del(userActivityFeedKey)
}

func (handlers *FeedHandlers) createUserActivityFeedCacheData(userID string, activities []entities.Activity) {
	userActivityIDs := [](string){}

	for _, activity := range activities {

		cacheActivityKey := fmt.Sprintf("activity_{}", activity.ID.Hex())
		activityBytes, err := json.Marshal(activity)
		if err != nil {
			return
		}
		activityString := string(activityBytes)
		handlers.cacheHelper.Set(cacheActivityKey, activityString, 0)

		userActivityIDs = append(userActivityIDs, activity.ID.Hex())
	}

	cacheUserActivityFeedKey := fmt.Sprintf("user_{}_activity_feed", userID)
	handlers.cacheHelper.Set(cacheUserActivityFeedKey, userActivityIDs, 0)
}
