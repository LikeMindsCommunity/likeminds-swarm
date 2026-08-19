package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/constants"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/enums"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/requests"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/responses"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/helpers"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/externalHelpers"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse pending post for response
func parsePendingPostResponse(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, pendingPost entities.PendingPost,
) responses.PostResponse {

	postResponse := parseSinglePostResponse(handlers, &pendingPost.PostData, loggedInUser)

	postResponse.MenuItems = getEntityMenuItems(handlers.cacheHelper, loggedInUser, constants.PendingPostEntityType,
		loggedInUser.UserId == pendingPost.UserId, pendingPost.PostData.IsPinned, pendingPost.PostData.IsHidden, pendingPost.UserId)

	postResponse.IsPendingPost = true
	postResponse.PostStatus = pendingPost.Status
	postResponse.IsDeleted = pendingPost.IsDeleted

	return postResponse
}

// Internal Method to parse multiple pending posts for response
func parseMultiplePendingPostResponse(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, pendingPosts []entities.PendingPost,
) []responses.PostResponse {

	response := []responses.PostResponse{}

	for _, pendingPost := range pendingPosts {
		response = append(response, parsePendingPostResponse(handlers, loggedInUser, pendingPost))
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
func fetchMultiplePendingPostsData(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, pendingPostIds []string,
) (map[string]responses.PostResponse, error) {

	// convert post_ids to object ids
	pendingPostObjectIds := helpers.ConvertIdsToObjectIds(pendingPostIds)

	// filter options to fetch posts from db
	filterOptions := gin.H{
		"_id": gin.H{
			"$in": pendingPostObjectIds,
		},
		"community_id": loggedInUser.CommunityId,
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
		postResponse[pendingPost.ID.Hex()] = parsePendingPostResponse(handlers, loggedInUser, pendingPost)
	}

	return postResponse, nil

}

func createPendingPostAfterValidation(handlers *FeedHandlers, userId string, communityId int, postRequest *requests.CreatePostRequest,
) (*entities.Post, error) {

	// Create pending post
	postId, err := handlers.pendingPostHelper.CreatePendingPostHelper(postRequest.Text, postRequest.Heading, communityId,
		userId, postRequest.Attachments, postRequest.ChatroomID, postRequest.TempID, postRequest.ParsedTopicIds, "",
		postRequest.Visibility, false, postRequest.IsAnonymous, postRequest.CreatedAt, enums.UnderReview, postRequest.UUIDs)
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
		enums.UnderReview, postRequest.UUIDs, postRequest.PostId)
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
	memberRole := headers[utils.HeadersMemberRole]
	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var cppr requests.CreatePostRequest
	if err := c.ShouldBindJSON(&cppr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             cppr.UserIsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
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
	postData, err := CreatePostAfterValidationFromType(handlers, userId, communityId, &cppr, headers)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch post response data
	pendingPostData, err := fetchPendingPost(handlers.pendingPostHelper, postData.ID.Hex(), communityId)
	if err == nil {
		response := gin.H{
			"post": parsePendingPostResponse(handlers, loggedInUser, *pendingPostData),
		}
		response = addMetadataInResponse(handlers, loggedInUser, response)

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
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
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
			"status":  arpr.Status,
			"post_id": "",
		},
	}

	var postData *entities.Post

	// If status is approved, call Create post API internally
	if arpr.Status == enums.Approved {

		if pendingPostData.PostId == "" {
			postData, err = createPostFromPendingPost(handlers, pendingPostData, headers)
			if err != nil {
				utils.GeneralAPIValidationError(c, err.Error())
				return
			}
		} else {
			postData, err = editPostFromPendingPost(handlers, communityId, &pendingPostData)
			if err != nil {
				utils.GeneralAPIValidationError(c, err.Error())
				return
			}
		}

		ctaData := gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     postData.ID.Hex(),
		}

		activityID, err := handlers.CreateActivity(communityId, []string{userId}, pendingPostData.UserId, constants.PostEntity,
			postData.ID, postData.UserId, constants.PendingPostAccepted, ctaData, false, false, primitive.NilObjectID, "")
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

		activityID, err := handlers.CreateActivity(communityId, []string{userId}, pendingPostData.UserId, constants.PendingPostEntity,
			pendingPostData.ID, pendingPostData.UserId, constants.PendingPostRejected, ctaData, false, false,
			primitive.NilObjectID, "")
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
		updateBody["$set"].(gin.H)["post_id"] = postData.ID.Hex()
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

func createPostFromPendingPost(handlers *FeedHandlers, pendingPostData entities.PendingPost, headers map[string]string) (*entities.Post, error) {

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
	postData, err := CreatePostAfterValidationFromType(handlers, pendingPostData.UserId, pendingPostData.CommunityID, &cpr, headers)
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
	memberRole := headers[utils.HeadersMemberRole]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editPendingPostRequest requests.EditPendingPostRequest
	if err := c.ShouldBindJSON(&editPendingPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             editPendingPostRequest.UserIsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
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

		topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, false, true, editPendingPostRequest.UserIsCm)
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

	// check the visibility of the pending post
	if editPendingPostRequest.Visibility == "" {
		editPendingPostRequest.Visibility = enums.PublicVisibility
	}

	if editPendingPostRequest.Visibility != enums.PrivateVisibility && editPendingPostRequest.Visibility != enums.PublicVisibility {
		utils.GeneralAPIValidationError(c, "Invalid visibility sent")
		return
	}

	isStatusChanged := false
	pendingPostStatus := pendingPostData.Status

	if pendingPostData.Status == enums.Rejected {
		isStatusChanged = true
		pendingPostStatus = enums.UnderReview
	}

	err = editPendingPostAfterValidation(handlers, communityId, userId, editPendingPostRequest.Attachments, editPendingPostRequest.Text,
		editPendingPostRequest.Heading, editPendingPostRequest.Visibility, editPendingPostRequest.UUIDs, pendingPostData,
		isStatusChanged, topicIDs, pendingPostStatus, "")

	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pending post response data
	pendingPostData, err = fetchPendingPost(handlers.pendingPostHelper, pendingPostId, communityId)
	if err == nil {
		response := gin.H{
			"post": parsePendingPostResponse(handlers, loggedInUser, *pendingPostData),
		}

		response = addMetadataInResponse(handlers, loggedInUser, response)

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
	memberRole := headers[utils.HeadersMemberRole]
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]

	isCm := utils.IsCMRole(memberRole)

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
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
		"post": parsePendingPostResponse(handlers, loggedInUser, *pendingPostData),
	}

	response = addMetadataInResponse(handlers, loggedInUser, response)

	// Generate success response
	utils.GenerateSuccessResponse(c, response)
}

// Exposed Method to fetch all the pending posts created by a User
func (handlers *FeedHandlers) FetchUserCreatedPendingPosts(c *gin.Context) {
	// fetch url params and headers
	headers := utils.GetHeaders(c)
	userId := c.Param("user_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]
	isCm := utils.IsCMRole(memberRole)

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
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

	pendingPostResponse := parseMultiplePendingPostResponse(handlers, loggedInUser, pendingPostResults)

	// response data
	response := gin.H{
		"posts":       pendingPostResponse,
		"total_count": pendingPostsResultsCount,
	}

	response = addMetadataInResponse(handlers, loggedInUser, response)

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
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
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
		"entity_type": constants.PendingPostEntity,
		"entity_id":   pendingPostData.ID,
	}
	err = handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)
	if err != nil {
		logging.Error("Error in deleting activity in pending post: ", err)
	}

	utils.GenerateSuccessResponse(c, response)
}

// Internal method to edit pending post after validation
func editPendingPostAfterValidation(handlers *FeedHandlers, communityId int, userId string, pendingPostAttachments []requests.AttachmentRequest,
	pendingPostText string, pendingPostHeading string, pendingPostVisibility string, pendingPostUUIDs []string, pendingPostData *entities.PendingPost,
	isStatusChanged bool, topicIDs []primitive.ObjectID, pendingPostStatus string, postId string,
) error {

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, constants.PostEntityType, pendingPostAttachments,
		pendingPostData.ID.Hex(), communityId, userId)
	if err != nil {
		return err
	}

	// update post data using helper method
	err = handlers.pendingPostHelper.EditPendingPostHelper(pendingPostData.ID, pendingPostText, pendingPostHeading, updatedAttachments,
		topicIDs, pendingPostVisibility, true, pendingPostStatus, pendingPostUUIDs, postId)
	if err != nil {
		return err
	}

	if isStatusChanged {
		// Send post for review
		err = SendPendingPostForReview(handlers, userId, communityId, pendingPostData.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Internal method to update the normal post from pending data
func editPostFromPendingPost(handlers *FeedHandlers, communityId int, pendingPostData *entities.PendingPost,
) (*entities.Post, error) {
	postId, _ := primitive.ObjectIDFromHex(pendingPostData.PostId)

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId.Hex(), communityId, true, []string{})
	if err != nil {
		return nil, err
	}

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

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, constants.PostEntityType, requestAttachments,
		pendingPostData.PostId, communityId, pendingPostData.UserId)
	if err != nil {
		return nil, err
	}

	// Update the post
	postData, err = editPostAfterValidation(handlers, communityId, postId, pendingPostData.PostData.Text, pendingPostData.PostData.Heading,
		updatedAttachments, helpers.ParseObjectIdsToStringArray(pendingPostData.PostData.TopicIds), postData.TopicIds,
		pendingPostData.PostData.Visibility)
	if err != nil {
		return nil, err
	}

	return postData, nil
}
