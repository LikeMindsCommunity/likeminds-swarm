package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse User activity list
func parseUserActivity(handler FeedHandlers, activities []entities.Activity,
	apiRevampV1Check bool, userId string) ([]interface{}, map[string]externalHelpers.MemberMeta, map[string]responses.TopicResponse, map[string]requests.WidgetResponse, error) {

	var postMetatadataValue string = externalHelpers.DefaultMetadataPostVariableValue
	var commentMetatadataValue string = externalHelpers.DefaultMetadataPostVariableValue

	response := []interface{}{}
	userDatas := map[string]externalHelpers.MemberMeta{}
	topicDatas := map[string]responses.TopicResponse{}
	widgetDatas := map[string]requests.WidgetResponse{}

	if len(activities) == 0 {
		return response, userDatas, topicDatas, widgetDatas, nil
	}

	postMetatadataValue = externalHelpers.GetPostVariableOrDefault(handler.cacheHelper, userId, activities[0].CommunityID)
	commentMetatadataValue = externalHelpers.GetCommentVariableOrDefault(handler.cacheHelper, userId, activities[0].CommunityID)

	userIds := []string{}
	topicIds := []primitive.ObjectID{}
	widgetIds := []primitive.ObjectID{}

	for _, activity := range activities {

		// Append last user from actionBy
		userIds = append(userIds, activity.ActionBy[len(activity.ActionBy)-1])
	}

	// Fetch Members Meta map
	success, userDatas := externalHelpers.FetchMemberMetaMap(userIds, userId, activities[0].CommunityID)
	if !success {
		return nil, nil, nil, nil, fmt.Errorf("error while fetching user meta")
	}

	for _, activity := range activities {

		activityUserUID := activity.ActionBy[len(activity.ActionBy)-1]
		activityUserData, exists := userDatas[activityUserUID]
		if !exists {
			continue
		}

		activityEntityData, err := getEntityData(handler, activity.EntityType, activity.EntityID, activity.CommunityID,
			apiRevampV1Check, userId, "")
		if err != nil {
			return response, userDatas, topicDatas, widgetDatas, err
		}

		activityText, err := getActivityText(activityUserData, activityEntityData, activity, postMetatadataValue, commentMetatadataValue)
		if err != nil {
			return response, userDatas, topicDatas, widgetDatas, err
		}

		activityCTA := getActivityCTA(handler, activity)

		userActivity := requests.UserActivityResponse{
			ID:                 activity.ID,
			ActionBy:           activity.ActionBy,
			ActionOn:           activity.ActionOn,
			EntityID:           activity.EntityID,
			EntityOwnerID:      activity.EntityOwnerID,
			UUID:               activity.EntityOwnerID,
			CTA:                activityCTA,
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
				Action:               enums.NewActivityActionFromInt(int(activity.Action), false),
			})

		} else {

			// Old User Activity Response
			response = append(response, requests.UserActivityResponseOld{
				UserActivityResponse: userActivity,
				EntityType:           activity.EntityType,
				Action:               activity.Action,
			})
		}

		// Parse topicIds and widgetIds from post
		if activity.EntityType == constants.Post {
			// Parse topicIds from postData
			if activityEntityData.(requests.PostResponse).Topics != nil {
				ids := activityEntityData.(requests.PostResponse).Topics
				topicIds = append(topicIds, ids...)
			}

			// Parse widgetIds from postData
			if activityEntityData.(requests.PostResponse).Attachments != nil {
				ids := getWidgetIdsFromAttachments(activityEntityData.(requests.PostResponse).Attachments)
				widgetIds = append(widgetIds, ids...)
			}
		}
	}

	// Parse topicsData from topicIds
	topicDatas, _ = fetchAndParseTopicsForResponse(handler.topicHelper, topicIds, activities[0].CommunityID)

	// Parse widgetsData from widgetIds
	widgetDatas, _ = parseWidgetsResponse(&handler, widgetIds, activities[0].CommunityID, enums.IsAdmin(userDatas[userId].State), userId)

	return response, userDatas, topicDatas, widgetDatas, nil
}

// Internal method to parse user profile activity list for uuid
func parseUserProfileActivity(handler FeedHandlers, activities []entities.Activity, apiRevampV1Check bool,
	uuid string, userId string) ([]interface{}, map[string]externalHelpers.MemberMeta, map[string]responses.TopicResponse,
	map[string]requests.WidgetResponse, error) {

	activitiesResponse, userDatas, topicsData, widgetsData := []interface{}{}, map[string]externalHelpers.MemberMeta{},
		map[string]responses.TopicResponse{}, map[string]requests.WidgetResponse{}

	if len(activities) == 0 {
		return activitiesResponse, userDatas, topicsData, widgetsData, nil
	}

	var postMetatadataValue string = externalHelpers.DefaultMetadataPostVariableValue
	var commentMetatadataValue string = externalHelpers.DefaultMetadataCommentVariableValue

	postMetatadataValue = externalHelpers.GetPostVariableOrDefault(handler.cacheHelper, userId, activities[0].CommunityID)
	commentMetatadataValue = externalHelpers.GetCommentVariableOrDefault(handler.cacheHelper, userId, activities[0].CommunityID)

	userIds := [](string){uuid}
	topicIds := []primitive.ObjectID{}
	widgetIds := []primitive.ObjectID{}

	for _, activity := range activities {
		// Append actionOn in userIds
		userIds = append(userIds, activity.ActionOn)
	}

	// Fetch Members Meta map
	success, userDatas := externalHelpers.FetchMemberMetaMap(userIds, userId, activities[0].CommunityID)
	if !success {
		return nil, nil, nil, nil, fmt.Errorf("error while fetching user meta")
	}

	for _, activity := range activities {

		actionByMetadata := activity.ActionByMetadata[uuid]

		// Update activity data
		activity.ActionBy = []string{uuid}
		activity.CreatedAt = actionByMetadata.CreatedAt
		activity.UpdatedAt = actionByMetadata.CreatedAt

		activityEntityData, postData, err := interface{}(nil), interface{}(nil), error(nil)

		// if action is comment on post, fetch comment data along with its post data
		switch activity.Action {
		case constants.CommentOnPost:
			activityEntityData, err = getEntityData(handler, constants.Comment, actionByMetadata.EntityId, activity.CommunityID,
				apiRevampV1Check, userId, activity.EntityID.Hex())
			if err != nil {
				continue
			}

			postData = *activityEntityData.(requests.FetchCommentResponse).Post

			// Update activity data
			activity.CTA = fmt.Sprintf(utils.CommentDetailRoute, activity.EntityID.Hex(), actionByMetadata.EntityId.Hex())
			activity.EntityID = actionByMetadata.EntityId
			activity.EntityType = constants.Comment
			activity.EntityOwnerID = uuid

		// Fetch entity data for other activities
		default:
			activityEntityData, err = getEntityData(handler, activity.EntityType, activity.EntityID, activity.CommunityID,
				apiRevampV1Check, userId, "")
			if err != nil {
				continue
			}

			postData = activityEntityData
		}

		activityText := getUserProfileActivityText(uuid, userId, activity.Action, userDatas, postMetatadataValue, commentMetatadataValue)

		// Make user activity response
		activityResponse := requests.UserActivityResponse{
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

			// API Revamp V1 activitiesResponse
			activitiesResponse = append(activitiesResponse, requests.UserActivityResponseV1{
				UserActivityResponse: activityResponse,
				EntityType:           enums.NewEntityTypeFromInt(int(activity.EntityType)),
				Action:               enums.NewActivityActionFromInt(int(activity.Action), true),
			})

		} else {

			// Old User Activity activitiesResponse
			activitiesResponse = append(activitiesResponse, requests.UserActivityResponseOld{
				UserActivityResponse: activityResponse,
				EntityType:           activity.EntityType,
				Action:               activity.Action,
			})
		}

		if postData != nil {

			// Parse topicIds from postData
			if postData.(requests.PostResponse).Topics != nil {
				ids := postData.(requests.PostResponse).Topics
				topicIds = append(topicIds, ids...)
			}

			// Parse widgetIds from postData
			if postData.(requests.PostResponse).Attachments != nil {
				ids := getWidgetIdsFromAttachments(postData.(requests.PostResponse).Attachments)
				widgetIds = append(widgetIds, ids...)
			}
		}

	}

	// Parse topicsData from topicIds
	topicsData, _ = fetchAndParseTopicsForResponse(handler.topicHelper, topicIds, activities[0].CommunityID)

	// Parse widgetsData from widgetIds
	widgetsData, _ = parseWidgetsResponse(&handler, widgetIds, activities[0].CommunityID, enums.IsAdmin(userDatas[userId].State), userId)

	return activitiesResponse, userDatas, topicsData, widgetsData, nil
}

func getActivityUserData(activity entities.Activity) (map[string]interface{}, string) {
	activityUserUID := activity.ActionBy[len(activity.ActionBy)-1]
	activityUserData := map[string]interface{}{}

	isSuccess, member_data := externalHelpers.FetchMemberMeta([]string{activityUserUID}, activity.ActionOn, activity.CommunityID)
	if !isSuccess || len(member_data.Members) == 0 {
		return nil, ""
	}

	activityUserData[activityUserUID] = member_data.Members[0]

	return activityUserData, activityUserUID
}

func getEntityData(handler FeedHandlers, entityType constants.EntityType, entityID primitive.ObjectID, communityID int,
	apiRevampV1Check bool, userId string, postIdForComment string) (interface{}, error) {

	switch entityType {
	case constants.Post:
		postData, err := fetchMultiplePostsData(&handler, []string{entityID.Hex()}, communityID, userId, false, "", "",
			apiRevampV1Check)
		if err != nil {
			return nil, err
		}

		return postData[entityID.Hex()], nil

	case constants.PendingPost:
		pendingPost, err := fetchPendingPost(handler.pendingPostHelper, entityID.Hex(), communityID)
		if err != nil {
			return nil, err
		}

		if pendingPost.Status != enums.Approved {
			pendingPostData, err := fetchMultiplePendingPostsData(&handler, []string{entityID.Hex()}, communityID, userId, false, "", "",
				apiRevampV1Check)
			if err != nil {
				return nil, err
			}

			return pendingPostData[entityID.Hex()], nil
		} else {
			postData, err := fetchMultiplePostsData(&handler, []string{pendingPost.NormalPostId}, communityID, userId, false, "", "",
				apiRevampV1Check)
			if err != nil {
				return nil, err
			}

			return postData[pendingPost.NormalPostId], nil
		}

	case constants.Comment:

		// If postIdForComment is not empty, fetch post data along with comment data
		if postIdForComment != "" {
			commentData, err := fetchCommentData(&handler, entityID.Hex(), postIdForComment, nil, userId, false,
				"", "", apiRevampV1Check, true, utils.DefaultRole)
			if err != nil {
				return nil, err
			}

			return commentData, nil

		} else {

			commentData, err := fetchMultipleCommentsData(&handler, []string{entityID.Hex()}, communityID, "", false, "", "",
				apiRevampV1Check, utils.DefaultRole)
			if err != nil {
				return nil, err
			}

			return commentData[entityID.Hex()].CommentResponse, nil
		}
	}

	return nil, nil
}

func getActivityText(activityByUserData externalHelpers.MemberMeta, activityEntityData interface{}, activity entities.Activity,
	postFeedMetadatValue string, commentFeedMetadaValue string) (string, error) {
	activityText := ""

	switch activity.Action {
	case constants.CreatePostPermitAdded:
		activityText += fmt.Sprintf("You now have the permission to create %s in the community. Start posting now.", utils.GetPluralOfString(postFeedMetadatValue))
		return activityText, nil

	case constants.CreatePostPermitRemoved:
		activityText += fmt.Sprintf("Your permission to create %s in the community has been removed.", utils.GetPluralOfString(postFeedMetadatValue))
		return activityText, nil

	case constants.CreateCommentPermitAdded:
		activityText += fmt.Sprintf("You now have the permission to add %s on the %s. Start engaging now.", utils.GetPluralOfString(commentFeedMetadaValue),
			utils.GetPluralOfString(postFeedMetadatValue))
		return activityText, nil

	case constants.CreateCommentPermitRemoved:
		activityText += fmt.Sprintf("Your permission to add %s and replies to the %s has been removed.", utils.GetPluralOfString(commentFeedMetadaValue),
			utils.GetPluralOfString(postFeedMetadatValue))
		return activityText, nil

	case constants.CMDeletedPost:
		activityText += fmt.Sprintf("Your %s has been deleted as it violates community guidelines. Reason: ", postFeedMetadatValue)
		activityText += activityEntityData.(requests.PostResponse).DeleteReason + "."
		return activityText, nil

	case constants.CMDeletedComment:
		activityText += "Your reply has been deleted as it violates community guidelines. Reason: "
		activityText += activityEntityData.(requests.CommentResponse).DeleteReason + "."
		return activityText, nil

	case constants.LikeOnPost:
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " liked your"

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.CommentOnPost:
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += fmt.Sprintf(" left a %s on your", commentFeedMetadaValue)

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.RepostOnPost:
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += " reposted your"

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.LikeOnComment:
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += fmt.Sprintf(" liked your %s", commentFeedMetadaValue)

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.CommentOnComment:
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += fmt.Sprintf(" replied on your %s", commentFeedMetadaValue)

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.TaggedInPost:
		activityText += getUserRoute(activityByUserData)

		activityText += " tagged you in their"

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.TaggedInPostComment:
		activityText += getUserRoute(activityByUserData)

		activityText += fmt.Sprintf(" tagged you in their %s", commentFeedMetadaValue)

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.AlsoCommentOnPost:
		activityText += getUserRoute(activityByUserData)
		activityText += getMultipleUserActivityText(activity)

		activityText += fmt.Sprintf(" also left a %s on", commentFeedMetadaValue)

		activityEntityOwnerUserData, activityEntityOwnerUserID := fetchActivityEntityOwnerUserData(activity)
		if activityEntityOwnerUserID != "" {
			activityText += " " + getUserRoute(activityEntityOwnerUserData[activityEntityOwnerUserID]) + "'s"
		}

		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil

	case constants.AcceptPendingPost:
		activityText += fmt.Sprintf("Your %s has been approved:", postFeedMetadatValue)
		activityText += getEntityText(activity.EntityType, activityEntityData, "")

		return activityText, nil

	case constants.RejectPendingPost:
		activityText += fmt.Sprintf("Your %s was not approved:", postFeedMetadatValue)
		activityText += getEntityText(activity.EntityType, activityEntityData, postFeedMetadatValue)

		return activityText, nil
	}

	return activityText, nil
}

func getActivityCTA(handlers FeedHandlers, activity entities.Activity) string {
	activityCTA := activity.CTA

	if activity.EntityType == constants.PendingPost && activity.Action != constants.RejectPendingPost {
		pendingPostData, _ := fetchPendingPost(handlers.pendingPostHelper, activity.EntityID.Hex(), activity.CommunityID)

		if pendingPostData.Status == enums.Approved {
			// CTA data for activity
			ctaData := gin.H{
				"entity_type": constants.PostEntityType,
				"post_id":     pendingPostData.NormalPostId,
			}
			activityCTA = parseCTAData(ctaData)
		}
	}

	return activityCTA
}

// Internal Method to fetch user profile activity text
func getUserProfileActivityText(uuid string, userId string, action constants.ActivityAction,
	userDatas map[string]externalHelpers.MemberMeta, postFeedMetadatValue string, commentMetatadataValue string) string {

	userRoute := getUserRoute(userDatas[uuid])

	switch action {
	case constants.CommentOnPost:
		return (userRoute + fmt.Sprintf(" left a %s on this ", commentMetatadataValue) + postFeedMetadatValue)

	case constants.LikeOnPost:
		return (userRoute + " liked this " + postFeedMetadatValue)

	default:
		return ""
	}

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
	userData := make(map[string]interface{})
	userID := activity.EntityOwnerID
	isSuccess := false

	isSuccess, userData[userID] = externalHelpers.FetchMemberMeta([]string{userID}, activity.ActionOn, activity.CommunityID)
	if !isSuccess {
		return nil, ""
	}
	if userData[userID] == nil {
		return nil, ""
	}

	userData[userID] = userData[userID].(*externalHelpers.MemberMetaResponse).Members[0]

	return userData, userID
}

func getEntityText(entityType constants.EntityType, activityEntityData interface{}, postFeedMetadatValue string) string {
	entityTextData := ""

	switch entityType {
	case constants.Post,
		constants.PendingPost:
		postResponse := activityEntityData.(requests.PostResponse)

		if postResponse.Heading != "" {
			entityTextData = activityEntityData.(requests.PostResponse).Heading
		} else if postResponse.Text != "" {
			entityTextData = activityEntityData.(requests.PostResponse).Text
		}

	case constants.Comment:
		entityTextData = activityEntityData.(requests.CommentResponse).Text
	}

	// if post text is nil, add attachment type as text
	if entityType == constants.Post && entityTextData == "" {
		return " " + getPostAttachmentType(activityEntityData.(requests.PostResponse)) + "."
	}

	if entityType == constants.Post && entityTextData != "" && postFeedMetadatValue != "" {
		return fmt.Sprintf(" %s \"", postFeedMetadatValue) + entityTextData + "\""
	} else if entityType == constants.Post && entityTextData != "" {
		return " \"" + entityTextData + "\""
	}

	if entityTextData == "" {

		if postFeedMetadatValue != "" {
			return postFeedMetadatValue + "."
		} else {
			return entityTextData + "."
		}
	}

	activityText := " \"" + entityTextData + "\""

	return activityText
}

func getPostAttachmentType(postResponse requests.PostResponse) string {
	if len(postResponse.Attachments) == 0 {
		return ""
	}

	intAttachmentType := postResponse.Attachments[0].AttachmentType
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
	entityID primitive.ObjectID, entityOwnerID string, action constants.ActivityAction, ctaData map[string]interface{},
	isRead bool, isDeleted bool, actionByEntityId primitive.ObjectID) (interface{}, error) {

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
		constants.AlsoCommentOnPost,
		constants.RepostOnPost,
		constants.AcceptPendingPost,
		constants.RejectPendingPost:

		activityID, err := handlers.activityHelper.CreateActivityHelper(communityID, actionBy, actionOn, entityType,
			entityID, entityOwnerID, action, cta, isRead, isDeleted, actionByEntityId)

		handlers.activityHelper.PushActivitytoCache(activityID)

		return activityID, err

	}

	return nil, nil
}

func (handlers *FeedHandlers) CreateAlsoCommentedActivity(activityID interface{}, postData *entities.Post,
	headers map[string]string, ctaData gin.H) {

	postCommentActivity, err := fetchActivity(handlers.activityHelper, activityID.(primitive.ObjectID).Hex())
	if err != nil {
		return
	}

	latestCommentUser := postCommentActivity.ActionBy[len(postCommentActivity.ActionBy)-1]
	previousCommentUsers := utils.RemoveAllOccurenceStringList(postCommentActivity.ActionBy, latestCommentUser)

	// if previousCommentUsers = [], no need to create activity
	if len(previousCommentUsers) == 0 {
		return
	}

	for _, previousCommentUser := range previousCommentUsers {

		// create also commented activity
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{latestCommentUser}, previousCommentUser,
			constants.Post, postData.ID, postData.UserId, constants.AlsoCommentOnPost, ctaData, false, false, primitive.NilObjectID)
		if err != nil {
			return
		}

		if activityID != nil {
			err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
			if err != nil {
				logging.Error("Failed to enqueue send notification : ", err)
			}
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
		constants.AlsoCommentOnPost,
		constants.RepostOnPost,
		constants.AcceptPendingPost,
		constants.RejectPendingPost:
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

			case constants.PendingPostEntityType:
				cta = fmt.Sprintf(utils.PendingPostDetailRoute, post_id)

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
	activityID, err := handlers.CreateActivity(community_id, []string{headers[utils.HeadersMemberId]}, user_id,
		constants.User, primitive.NilObjectID, user_id, action, gin.H{}, false, false, primitive.NilObjectID)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activityID != nil {
		err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
		if err != nil {
			logging.Error("Failed to enqueue send notification : ", err)
		}
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
	userID := headers[utils.HeadersMemberId]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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
	err = handlers.activityHelper.UpdateActivityByIDHelper(activity[0].ID, activityUpdateData, true, false)
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
		"is_deleted":   false,
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

// FetchUserProfileActivity | method to Fetch User Profile Activity for uuid
func (handlers *FeedHandlers) FetchUserProfileActivity(c *gin.Context) {

	// fetch url params and headers
	headers := utils.GetHeaders(c)
	uuid := c.Param("user_id")

	// api revamp v1 check
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityID := externalHelpers.GetCommunityId(c)
	if communityID == externalHelpers.DefaultCommunityId {
		return
	}

	// Filter all activities where uuid is present in action_by_metadata
	actionByMetaUserKey := fmt.Sprintf("action_by_metadata.%s", uuid)

	// activity filter data
	activityFilterData := gin.H{
		"action": gin.H{
			"$in": []int{6, 7}, // Only fetch LikeOnPost and CommentOnPost activities
		},
		actionByMetaUserKey: gin.H{
			"$exists": true,
		},
		"community_id": communityID,
		"is_deleted":   false,
	}

	// Sort on created_at present in action_by_metadata
	activitySortKey := fmt.Sprintf(actionByMetaUserKey + ".created_at")

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
	activityResponse, userDatas, topicDatas, widgetDatas, err := parseUserProfileActivity(*handlers, activityResults, apiRevampV1Check,
		uuid, headers[utils.HeadersMemberId])
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
