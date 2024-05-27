package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse pending post for response
func parsePendingPostResponse(handlers *FeedHandlers, pendingPost entities.PendingPost, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool, memberRole string,
) responses.PostResponse {

	memberRole = utils.GuestRole

	loggedInUser := LoggedInUserParams{
		UserId:           userId,
		IsCm:             isCm,
		PlatformCode:     platformCode,
		VersionCode:      versionCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	postResponse := parseSinglePostResponse(handlers, &pendingPost.PostData, &loggedInUser)

	postResponse.MenuItems = getEntityMenuItems(constants.PendingPostEntityType, isCm,
		userId == pendingPost.UserId, pendingPost.PostData.IsPinned, versionCode, platformCode, userId, pendingPost.PostData.CommunityId, handlers.cacheHelper)

	postResponse.IsPendingPost = true
	postResponse.PostStatus = pendingPost.Status
	postResponse.IsDeleted = pendingPost.IsDeleted

	return postResponse
}

// Internal Method to parse multiple pending posts for response
func parseMultiplePendingPostResponse(handlers *FeedHandlers, pendingPosts []entities.PendingPost, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool, memberRole string,
) []responses.PostResponse {

	response := []responses.PostResponse{}

	for _, pendingPost := range pendingPosts {
		response = append(response, parsePendingPostResponse(handlers, pendingPost, userId, isCm, versionCode, platformCode, apiRevampV1Check, memberRole))
	}

	return response
}

// Internal Method to fetch pending post using post_id and community_id
func fetchPendingPost(helper interfaces.PendingPostHelper, pendingPostId string, communityId int) (*entities.PendingPost, error) {

	// filter data
	filterData := gin.H{
		"_id":          pendingPostId,
		"is_deleted":   false,
		"community_id": communityId,
	}

	// fetch post using helper method
	results, err := helper.FindPendingPostHelper(filterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of post_id
	if len(results) == 0 {
		return nil, fmt.Errorf("invalid pending_post_id")
	}

	return &results[0], nil
}

// Internal Method to fetch multiple posts data using post_ids
func fetchMultiplePendingPostsData(handlers *FeedHandlers, pendingPostIds []string, communityId int, userId string,
	isCm bool, versionCode string, platformCode string, apiRevampV1Check bool) (map[string]responses.PostResponse, error) {

	// convert post_ids to object ids
	pendingPostObjectIds := helpers.ConvertIdsToObjectIds(pendingPostIds)

	// filter options to fetch posts from db
	filterOptions := gin.H{
		"_id": gin.H{
			"$in": pendingPostObjectIds,
		},
		"community_id": communityId,
	}

	// fetch posts using helper method
	pendingPostLists, err := handlers.pendingPostHelper.FindPendingPostHelper(filterOptions, gin.H{})
	if err != nil {
		return nil, err
	}

	// Make key value pair of post_id -> PostResponse
	postResponse := map[string]responses.PostResponse{}

	// parse post data from pending posts
	for _, pendingPost := range pendingPostLists {
		postResponse[pendingPost.ID.Hex()] = parsePendingPostResponse(handlers, pendingPost, userId, isCm, versionCode,
			platformCode, apiRevampV1Check, utils.DefaultRole)
	}

	return postResponse, nil

}

func createPendingPostAfterValidation(handlers *FeedHandlers, userId string, communityId int,
	postRequest *requests.CreatePostRequest) (*entities.Post, error) {

	// Create pending post
	postId, err := handlers.pendingPostHelper.CreatePendingPostHelper(postRequest.Text, postRequest.Heading, communityId,
		userId, postRequest.Attachments, postRequest.ChatroomID, postRequest.TempID, postRequest.ParsedTopicIds, "",
		postRequest.Visibility, false, postRequest.CreatedAt, enums.UnderReview, postRequest.UUIDs)
	if err != nil {
		return nil, err
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, postRequest.PostType, postRequest.Attachments,
		postId.(primitive.ObjectID).Hex(), communityId, userId)
	if err != nil {
		return nil, err
	}

	err = handlers.pendingPostHelper.EditPendingPostHelper(postId.(primitive.ObjectID), postRequest.Text,
		postRequest.Heading, updatedAttachments, postRequest.ParsedTopicIds, postRequest.Visibility, false,
		enums.UnderReview, postRequest.UUIDs)
	if err != nil {
		return nil, err
	}

	// fetch pending post data using new post_id
	pendingPostData, err := fetchPendingPost(handlers.pendingPostHelper, postId.(primitive.ObjectID).Hex(), communityId)
	if err != nil {
		return nil, err
	}

	err = SendPendingPostForReview(handlers, userId, communityId, postId.(primitive.ObjectID))
	if err != nil {
		return nil, err
	}

	return &pendingPostData.PostData, nil
}

func SendPendingPostForReview(handlers *FeedHandlers, userId string, communityId int, pendingPostId primitive.ObjectID) error {
	// Call caravan API to create a review report for the pending post
	reportId, err := externalHelpers.CreatePendingPostReport(userId, communityId, pendingPostId.Hex())
	if err != nil {

		// Delete the pending post if there is an error in sending the report
		err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostId,
			gin.H{"$set": gin.H{"is_deleted": true}})
		if err != nil {
			// Log the error
			logging.Error(fmt.Sprint(
				"Error in deleting the pending post: ", pendingPostId.Hex(), " after error in sending for review: ", err.Error()))
		}

		return err
	} else {
		// Update the report_id in pending post if there is no error in sending the report
		err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostId,
			gin.H{"$set": gin.H{"report_id": reportId}})
		if err != nil {
			// Log the error
			logging.Error(fmt.Sprint(
				"Error in updating the pending post: ", pendingPostId.Hex(), " after error in sending for review: ", err.Error()))
		}

		return err
	}
}

// Exposed method to create a pending post for review (similar to Create post method)
func (handlers *FeedHandlers) CreatePendingPostForReview(c *gin.Context) {

	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	memberRole := headers[utils.HeaderMemberRole]
	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var cppr requests.CreatePostRequest
	if err := c.ShouldBindJSON(&cppr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Validate create post request
	errorMeta, err := validateCreatePostRequest(handlers, userId, communityId, apiRevampV1Check, &cppr)
	if err != nil && errorMeta == nil {
		// if errorMeta is nil, then it is a general validation error
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create pending post using internal method
	cppr.PostType = constants.PendingPostEntityType
	postData, err := createPostAfterValidation(handlers, userId, communityId, &cppr, headers)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch post response data
	pendingPostData, err := fetchPendingPost(handlers.pendingPostHelper, postData.ID.Hex(), communityId)
	if err == nil {
		response := gin.H{
			"post": parsePendingPostResponse(handlers, *pendingPostData, headers[utils.HeadersMemberId], cppr.UserIsCm,
				versionCode, platformCode, apiRevampV1Check, memberRole),
		}
		response = addMetadataInResponse(handlers, response, communityId, userId, platformCode, versionCode, cppr.UserIsCm,
			apiRevampV1Check)

		// Generate success response
		utils.GenerateSuccessResponse(c, response)
	} else {
		utils.GeneralAPIValidationError(c, utils.PendingPostCreationError)
	}
}

// Exposed method to approve/reject a pending post under review
func (handlers *FeedHandlers) ApproveOrRejectPendingPost(c *gin.Context) {

	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	pendingPostId := c.Param("pending_post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var arpr requests.ApproveRejectPendingPostRequest
	if err := c.ShouldBindJSON(&arpr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of status
	if arpr.Status != enums.Approved && arpr.Status != enums.Rejected {
		utils.GeneralAPIValidationError(c, "Invalid status sent")
		return
	}

	// Get pending post data
	filterData := gin.H{
		"_id":          pendingPostId,
		"community_id": communityId,
		"status":       enums.UnderReview,
		"is_deleted":   false,
	}

	pendingPostsData, err := handlers.pendingPostHelper.FindPendingPostHelper(filterData, nil)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(pendingPostsData) == 0 {
		utils.GeneralAPIValidationError(c, "Invalid pending post id sent")
		return
	}

	pendingPostData := pendingPostsData[0]

	updateBody := gin.H{
		"$set": gin.H{
			"status":         arpr.Status,
			"normal_post_id": "",
		},
	}

	var postData *entities.Post

	// If status is approved, call Create post API internally
	if arpr.Status == enums.Approved {

		postData, err = createNormalPostFromPendingPost(handlers, pendingPostData, headers)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		ctaData := gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     postData.ID.Hex(),
		}

		activityID, err := handlers.CreateActivity(communityId, []string{userId}, pendingPostData.UserId, constants.Post,
			postData.ID, postData.UserId, constants.PendingPostAccepted, ctaData, false, false, primitive.NilObjectID)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			// Send Approval notification
			sendPendingPostApprovalNotification(*handlers, postData.UserId, communityId, postData.ID.Hex())
		}

	} else {
		ctaData := gin.H{
			"entity_type": constants.PendingPostEntityType,
			"post_id":     pendingPostData.ID.Hex(),
		}

		activityID, err := handlers.CreateActivity(communityId, []string{userId}, pendingPostData.UserId, constants.PendingPost,
			pendingPostData.ID, pendingPostData.UserId, constants.PendingPostRejected, ctaData, false, false, primitive.NilObjectID)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			// Send Rejection notification
			sendPendingPostRejectionNotification(*handlers, pendingPostData.UserId, communityId, pendingPostData.ID.Hex())
		}
	}

	if postData != nil {
		updateBody["$set"].(gin.H)["normal_post_id"] = postData.ID
	}

	// Update status of pending post
	err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostData.ID, updateBody)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Generate success response
	utils.GenerateSuccessResponse(c, nil)

}

func createNormalPostFromPendingPost(handlers *FeedHandlers, pendingPostData entities.PendingPost, headers map[string]string) (*entities.Post, error) {

	// Create attachments
	requestAttachments := []requests.AttachmentRequest{}

	// marshal attachments
	bytes, err := json.Marshal(pendingPostData.PostData.Attachments)
	if err != nil {
		return nil, err
	}

	// Unmarshall to Request attachments
	err = json.Unmarshal(bytes, &requestAttachments)
	if err != nil {
		return nil, err
	}

	cpr := requests.CreatePostRequest{
		Text:           pendingPostData.PostData.Text,
		Heading:        pendingPostData.PostData.Heading,
		Attachments:    requestAttachments,
		ChatroomID:     pendingPostData.PostData.ChatroomId,
		ParsedTopicIds: pendingPostData.PostData.TopicIds,
		OriginalAuthor: pendingPostData.PostData.OriginalAuthorUUID,
		UUIDs:          pendingPostData.UUIDs,
		Visibility:     pendingPostData.PostData.Visibility,
		TempID:         pendingPostData.PostData.TempId,
		IsRepost:       pendingPostData.PostData.IsRepost,
	}

	// create post using internal method
	cpr.PostType = constants.PostEntityType
	postData, err := createPostAfterValidation(handlers, pendingPostData.UserId, pendingPostData.CommunityID, &cpr, headers)
	if err != nil {
		return nil, err
	}

	// update all the widgets for the newly created post
	updateWidgetsForNewlyCreatePostFromPendingPost(handlers, postData.ID.Hex(), pendingPostData.CommunityID, pendingPostData.PostData.Attachments)

	return postData, nil
}

// Internal method to update all the widgets for the newly created post
func updateWidgetsForNewlyCreatePostFromPendingPost(handlers *FeedHandlers, postId string, communityId int,
	attachments []entities.Attachment) error {

	for _, attachment := range attachments {

		switch attachment.AttachmentType {
		case enums.CustomWidget, enums.PollWidget, enums.ArticleWidget:

			createdByLm := true

			if attachment.AttachmentType == enums.CustomWidget {
				createdByLm = false
			}

			if attachment.AttachmentMeta.EntityID != primitive.NilObjectID {

				_, err := editWidget(handlers, attachment.AttachmentMeta.EntityID.Hex(), postId, constants.PostEntityType, createdByLm, nil, nil, communityId)
				if err != nil {
					fmt.Println("Error in updating widget for post: ", postId, " and widget: ", attachment.AttachmentMeta.EntityID.Hex())
				}
			}
		}

	}

	return nil
}

// Exposed Method to edit a Post
func (handlers *FeedHandlers) EditPendingPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]
	pendingPostId := c.Param("pending_post_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	memberRole := headers[utils.HeaderMemberRole]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editPendingPostRequest requests.EditPendingPostRequest
	if err := c.ShouldBindJSON(&editPendingPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pending post data
	pendingPostData, err := fetchPendingPost(handlers.pendingPostHelper, pendingPostId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Check if the pending post id is already approved or not
	if pendingPostData.Status == enums.Approved {
		utils.GeneralAPIValidationError(c, "Cannot update an approved post")
		return
	}

	// Check if the pending post id is already approved or not
	if pendingPostData.IsDeleted {
		utils.GeneralAPIValidationError(c, "Cannot update a deleted post")
		return
	}

	// Check if user is post creator
	if memberRole != utils.CMRole && pendingPostData.UserId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// validation of attachment objects
	err = ValidateAndUpdateAttachments(handlers, communityId, enums.EntityTypePendingPost, editPendingPostRequest.Attachments, apiRevampV1Check,
		true, pendingPostData.PostData.IsRepost)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validates a respost's post attachement in edit request
	if pendingPostData.PostData.IsRepost && !validateRepostPostAttachment(&pendingPostData.PostData, editPendingPostRequest.EditPostRequest) {
		utils.GeneralAPIValidationError(c, "Cannot update repost's post attachment")
		return
	}

	// If NSFW Filtering is enabled & attachments are present, check for NSFW content
	if len(editPendingPostRequest.Attachments) > 0 {
		errorMeta, err := validateAndUpdatePostImagesForNSFWContent(handlers.cacheHelper, headers[utils.HeadersMemberId], communityId,
			&editPendingPostRequest.Attachments, &pendingPostData.PostData.Attachments)
		if errorMeta != nil {
			utils.CustomAPIErrorWithMeta(c, http.StatusBadRequest, err.Error(), errorMeta)
			return
		}
	}

	// strip text and check if it is empty
	editPendingPostRequest.Text = strings.TrimSpace(editPendingPostRequest.Text)

	if editPendingPostRequest.Text == "" && editPendingPostRequest.Heading == "" && len(editPendingPostRequest.Attachments) == 0 {
		utils.GeneralAPIValidationError(c, "Can't Edit post without content")
		return
	}

	topicIDs := pendingPostData.PostData.TopicIds

	// fetch all the topics sent in the edit post body
	if editPendingPostRequest.TopicIds != nil {
		// convert topic_ids to object ids
		topicIDs = helpers.ConvertIdsToObjectIds(editPendingPostRequest.TopicIds)

		topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, false)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		// Validation of Topics
		if len(topics) != len(topicIDs) {
			utils.GeneralAPIValidationError(c, "Invalid topic_ids sent")
			return
		}
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, constants.PostEntityType, editPendingPostRequest.Attachments,
		pendingPostId, communityId, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// check the visibility of the pending post
	if editPendingPostRequest.Visibility == "" {
		editPendingPostRequest.Visibility = enums.PublicVisibility
	}

	if editPendingPostRequest.Visibility != enums.PrivateVisibility && editPendingPostRequest.Visibility != enums.PublicVisibility {
		utils.GeneralAPIValidationError(c, "Invalid visibility sent")
		return
	}

	isStatusChanged := false
	pendingPosStatus := pendingPostData.Status

	if pendingPostData.Status == enums.Rejected {
		isStatusChanged = true
		pendingPosStatus = enums.UnderReview
	}

	// update post data using helper method
	err = handlers.pendingPostHelper.EditPendingPostHelper(pendingPostData.ID, editPendingPostRequest.Text, editPendingPostRequest.Heading, updatedAttachments,
		topicIDs, editPendingPostRequest.Visibility, true, pendingPosStatus, editPendingPostRequest.UUIDs)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if isStatusChanged {
		// Send post for review
		err = SendPendingPostForReview(handlers, userId, communityId, pendingPostData.ID)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// fetch pending post response data
	pendingPostData, err = fetchPendingPost(handlers.pendingPostHelper, pendingPostId, communityId)
	if err == nil {
		response := gin.H{
			"post": parsePendingPostResponse(handlers, *pendingPostData, headers[utils.HeadersMemberId], false,
				versionCode, platformCode, apiRevampV1Check, memberRole),
		}

		response = addMetadataInResponse(handlers, response, communityId, userId, platformCode, versionCode, false,
			apiRevampV1Check)

		// Generate success response
		utils.GenerateSuccessResponse(c, response)
	} else {
		utils.GeneralAPIValidationError(c, utils.PendingPostUpdationError)
	}
}

// Exposed Method to fetch a pending post using pending_post_id
func (handlers *FeedHandlers) FetchPendingPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	pendingPostId := c.Param("pending_post_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	userId := headers[utils.HeadersMemberId]
	memberRole := headers[utils.HeaderMemberRole]
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch pending post data
	pendingPostData, err := fetchPendingPost(handlers.pendingPostHelper, pendingPostId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	if pendingPostData.UserId != userId && memberRole != utils.CMRole {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	response := gin.H{
		"post": parsePendingPostResponse(handlers, *pendingPostData, headers[utils.HeadersMemberId], false,
			versionCode, platformCode, apiRevampV1Check, memberRole),
	}

	response = addMetadataInResponse(handlers, response, communityId, userId, platformCode, versionCode, false,
		apiRevampV1Check)

	// Generate success response
	utils.GenerateSuccessResponse(c, response)
}

// Exposed Method to fetch all the pending posts created by a User
func (handlers *FeedHandlers) FetchUserCreatedPendingPosts(c *gin.Context) {
	// fetch url params and headers
	headers := utils.GetHeaders(c)
	userId := c.Param("user_id")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// post filter data
	postFilterData := gin.H{
		"user_id":      userId,
		"is_deleted":   false,
		"community_id": communityId,
		"status": gin.H{
			"$in": []string{
				enums.UnderReview,
				enums.Rejected,
			},
		},
	}

	// filter options
	postFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pending post using helper method
	pendingPostResults, err := handlers.pendingPostHelper.FindPendingPostHelper(postFilterData, postFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	pendingPostsResultsCount, err := handlers.pendingPostHelper.CountPendingPostHelper(postFilterData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	pendingPostResponse := parseMultiplePendingPostResponse(handlers, pendingPostResults, userId, isCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, utils.DefaultRole)

	// response data
	response := gin.H{
		"posts":       pendingPostResponse,
		"total_count": pendingPostsResultsCount,
	}

	response = addMetadataInResponse(handlers, response, communityId, userId, platformCode, versionCode, false,
		apiRevampV1Check)

	// return final response
	utils.GenerateSuccessResponse(c, response)
}

// Exposed Method to delete a pending post
func (handlers *FeedHandlers) DeletePendingPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	pendingPostId := c.Param("pending_post_id")

	userId := headers[utils.HeadersMemberId]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deletePostRequest requests.DeletePostRequest
	if err := c.ShouldBindJSON(&deletePostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pending post using helper method
	pendingPostData, err := fetchPendingPost(handlers.pendingPostHelper, pendingPostId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if userId != pendingPostData.UserId {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// update data
	updateData := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": deletePostRequest.DeleteReason,
			"deleted_by":    userId,
		},
	}

	// update post using the helper method
	err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostData.ID, updateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	response := gin.H{
		"report_id": pendingPostData.ReportID,
	}

	// remove activity for the pending post
	deleteActivityFilter := gin.H{
		"entity_type": constants.PendingPost,
		"entity_id":   pendingPostData.ID,
	}
	err = handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)
	if err != nil {
		logging.Error("Error in deleting activity in pending post: ", err)
	}

	utils.GenerateSuccessResponse(c, response)
}
