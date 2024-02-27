package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to fetch comments count of a Post
func fetchPostCommentsCount(helper interfaces.CommentHelper, postId string) (int64, error) {
	// comment filter data
	commentFilterData := gin.H{
		"post_id":    postId,
		"is_deleted": false,
		"level":      constants.CommentBaseLevel,
	}

	// fetch likes count using helper method
	likesCount, err := helper.CountCommentHelper(commentFilterData)
	if err != nil {
		return 0, err
	}

	return likesCount, nil
}

// Internal Method to fetch replies count of a Comment
func fetchCommentRepliesCount(helper interfaces.CommentHelper, commentId string) (int64, error) {
	commentData, err := fetchCommentByIdInternal(helper, commentId)
	if err != nil {
		return 0, err
	}

	replyFilterData := gin.H{
		"_id": gin.H{
			"$in": commentData.Replies,
		},
		"is_deleted": false,
	}

	// fetch replies count using helper method
	likesCount, err := helper.CountCommentHelper(replyFilterData)
	if err != nil {
		return 0, err
	}

	return likesCount, nil
}

// Internal Method to fetch parent comment of a Comment
func fetchParentComment(helper interfaces.CommentHelper, commentId primitive.ObjectID,
	postId primitive.ObjectID) (*entities.Comment, error) {
	// comment filter data
	commentFilterData := gin.H{
		"replies":    commentId,
		"is_deleted": false,
		"post_id":    postId,
	}

	// fetch comment using helper method
	commentResults, err := helper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &commentResults[0], nil
}

// Internal Method to fetch a comment using comment_id
func fetchCommentByIdInternal(helper interfaces.CommentHelper, commentId string) (*entities.Comment, error) {
	// comment filter data
	commentFilterData := gin.H{
		"_id":        commentId,
		"is_deleted": false,
	}

	// fetch comment using helper method
	commentResults, err := helper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &commentResults[0], nil
}

// fetchCommentByID | get comment by id
func fetchCommentByID(helper interfaces.CommentHelper, commentId string) (*entities.Comment, error) {
	filter := gin.H{
		"_id": commentId,
	}

	commentResults, err := helper.FindCommentHelper(filter, gin.H{})
	if err != nil {
		return nil, err
	}

	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id")
	}

	return &commentResults[0], nil
}

// Internal Method to fetch multiple comments data using comment_ids
func fetchMultipleCommentsData(handlers *FeedHandlers, commentIds []string, communityId int,
	userId string, isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool) (map[string]requests.CommentWithParentResponse, error) {

	// convert comment_ids to object ids
	commentObjectIds := helpers.ConvertIdsToObjectIds(commentIds)

	// comment filter data
	commentFilterData := gin.H{
		"_id": gin.H{
			"$in": commentObjectIds,
		},
		"community_id": communityId,
	}

	// fetch comments using helper method
	comments, err := handlers.commentHelper.FindCommentHelper(commentFilterData, nil)
	if err != nil {
		return nil, err
	}

	// Make key value pair map for response, comment_id -> comment
	parsedCommentsResponse := map[string]requests.CommentWithParentResponse{}
	for _, comment := range comments {

		// Parse comment for response
		parseCommentsResponse := parseCommentWithParentResponse(handlers, comment, userId, isCm, versionCode, platformCode, apiRevampV1Check)

		// Add to response map
		parsedCommentsResponse[comment.ID.Hex()] = parseCommentsResponse

	}

	return parsedCommentsResponse, nil
}

// Internal Method to fetch a comment using comment_id and post_id
func fetchComment(helper interfaces.CommentHelper, commentId string, postId string) (*entities.Comment, error) {
	// comment filter data
	commentFilterData := gin.H{
		"_id":        commentId,
		"is_deleted": false,
		"post_id":    postId,
	}

	// fetch comment using helper method
	commentResults, err := helper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &commentResults[0], nil
}

// Internal Method to parse comment for response
func parseCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comment entities.Comment, userId string, isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool, cacheHelper cache.Helper) requests.CommentResponse {
	likesCount, _ := fetchEntityLikesCount(likeHelper, comment.ID.Hex(), constants.CommentEntityType)
	var response requests.CommentResponse

	response.ID = comment.ID
	response.TempID = comment.TempID
	response.Text = comment.Text
	response.Level = comment.Level
	response.CommunityId = comment.CommunityId
	response.PostId = comment.PostId
	response.UserId = comment.UserId
	response.UUID = comment.UserId
	response.IsLiked = fetchUserLikedStatusByEntity(likeHelper, comment.ID.Hex(), constants.CommentEntityType, userId)
	response.LikesCount = int(likesCount)
	response.IsDeleted = comment.IsDeleted
	response.IsEdited = comment.IsEdited
	response.MenuItems = getEntityMenuItems(constants.CommentEntityType, isCm, userId == comment.UserId, false, versionCode, platformCode, userId, comment.CommunityId, cacheHelper)

	if comment.Level == constants.CommentBaseLevel {
		repliesCount, _ := fetchCommentRepliesCount(commentHelper, comment.ID.Hex())
		response.CommentsCount = int(repliesCount)
	}

	if comment.IsDeleted {
		response.DeleteReason = comment.DeleteReason
		response.DeletedBy = comment.DeletedBy
		response.DeletedByUUID = comment.DeletedBy
	}

	response.CreatedAt = int(comment.CreatedAt.UnixMilli())
	response.UpdatedAt = int(comment.UpdatedAt.UnixMilli())

	// ApiRevampV1Check to remove user_id and community_id from comment
	if apiRevampV1Check {
		response.UserId = ""
		response.CommunityId = 0
	}

	return response
}

// Method to fetch comment response using comment id
func FetchSingleCommentWithParentResponse(handlers *FeedHandlers, commentId string) (*requests.CommentWithParentResponse, error) {

	comment, err := fetchCommentByID(handlers.commentHelper, commentId)
	if err != nil {
		return nil, err
	}

	response := parseCommentWithParentResponse(handlers, *comment, comment.UserId, false, "", "", false)

	return &response, nil
}

// Internal Method to parse multiple comments for response
func parseMultipleCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comments []entities.Comment, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool, cacheHelper cache.Helper) []requests.CommentResponse {
	var response []requests.CommentResponse
	for _, comment := range comments {
		response = append(response, parseCommentResponse(likeHelper, commentHelper, comment, userId, isCm, versionCode, platformCode,
			apiRevampV1Check, cacheHelper))
	}

	return response
}

// Internal method to parse comments for FetchCommentsResponse
func parseCommentWithParentResponse(handlers *FeedHandlers, comment entities.Comment, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool) requests.CommentWithParentResponse {

	fetchCommentResponse := requests.CommentWithParentResponse{
		CommentResponse: parseCommentResponse(handlers.likeHelper, handlers.commentHelper, comment, userId, isCm, versionCode, platformCode,
			apiRevampV1Check, handlers.cacheHelper),
	}

	// Fetch parent comment if exists
	if fetchCommentResponse.Level > constants.CommentBaseLevel {
		parentComment, err := fetchParentComment(handlers.commentHelper, comment.ID, comment.PostId)
		if err == nil {
			parentCommentResponse := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *parentComment, userId, isCm, versionCode, platformCode,
				apiRevampV1Check, handlers.cacheHelper)
			fetchCommentResponse.ParentComment = &parentCommentResponse
		}
	}

	return fetchCommentResponse

}

// Internal Method to parse comment data for FetchComment API
func parseFetchCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	rawComment *entities.Comment, replies []requests.CommentResponse, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool, cacheHelper cache.Helper) requests.FetchCommentResponse {
	var response requests.FetchCommentResponse

	response.CommentResponse = parseCommentResponse(likeHelper, commentHelper, *rawComment, userId, isCm, versionCode, platformCode,
		apiRevampV1Check, cacheHelper)

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []requests.CommentResponse{}
	}

	if response.CommentResponse.Level > constants.CommentBaseLevel {
		commentData, err := fetchParentComment(commentHelper, rawComment.ID, rawComment.PostId)
		if err == nil {
			parentCommentResponse := parseCommentResponse(likeHelper, commentHelper, *commentData, userId, isCm, versionCode, platformCode,
				apiRevampV1Check, cacheHelper)
			response.ParentComment = &parentCommentResponse
		}
	}

	return response
}

// Internal Method to fetch comment data with postId
func fetchCommentData(handlers *FeedHandlers, commentId string, postId string, filterOptions map[string]interface{},
	memberId string, isCm bool, versionCode string, platformCode string, apiRevampV1Check bool, getPostData bool) (interface{}, error) {

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId)
	if err != nil {
		return nil, err
	}

	commentFilterData := gin.H{
		"_id": gin.H{
			"$in": commentData.Replies,
		},
		"is_deleted": false,
		"post_id":    postId,
	}

	// fetch comment replies using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, filterOptions)
	if err != nil {
		return nil, err
	}

	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, memberId, isCm,
		versionCode, platformCode, apiRevampV1Check, handlers.cacheHelper)
	fetchCommentResponse := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentData, repliesResponse, memberId, isCm, versionCode, platformCode, apiRevampV1Check, handlers.cacheHelper)

	// fetch post data if getPostData is true
	if getPostData {
		postData, err := fetchPost(handlers.postHelper, postId, commentData.CommunityId)
		if err != nil {
			return nil, err
		}

		// Parse post response and append to Comment's post_data
		parsePostResponse := parsePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper, handlers.topicHelper, handlers.widgetHelper,
			*postData, memberId, isCm, versionCode, platformCode, apiRevampV1Check, handlers.cacheHelper)
		fetchCommentResponse.Post = &parsePostResponse
	}
	return fetchCommentResponse, nil
}

// Exposed Method to fetch comment by comment_id
func (handlers *FeedHandlers) FetchCommentById(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	commentId := c.Param("comment_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	if paramIsCm == "true" {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch comment data
	commentData, err := fetchCommentByIdInternal(handlers.commentHelper, commentId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	commentFilterData := gin.H{
		"_id": gin.H{
			"$in": commentData.Replies,
		},
		"is_deleted": false,
		"post_id":    commentData.PostId.Hex(),
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, commentFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, headers[utils.HeadersMemberId],
		isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, handlers.cacheHelper)
	fetchCommentResponse := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentData, repliesResponse, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check, handlers.cacheHelper)

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"comment": fetchCommentResponse,
	})
}

// Exposed Method to fetch multiple comments by comment_ids
func (handlers *FeedHandlers) FetchComments(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Get Query Params
	paramCommentIds := c.Query("comment_ids")
	paramIsCm, _ := strconv.ParseBool(c.Query("user_is_cm"))

	// Check if user is CM or not
	if !paramIsCm {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Unmarshal comment_ids
	var commentIds []string
	err := json.Unmarshal([]byte(paramCommentIds), &commentIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Fetch comments using comment_ids
	comments, err := fetchMultipleCommentsData(handlers, commentIds, communityId, headers[utils.HeadersMemberId], paramIsCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// response data
	response := gin.H{
		"success":  true,
		"comments": comments,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed method to fetch comment by comment_id and post_id
func (handlers *FeedHandlers) FetchComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	if paramIsCm == "true" {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post data
	_, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, commentId, postId, commentFilterOptions, headers[utils.HeadersMemberId], isCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, false)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response["comment"] = fetchCommentResponse

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to comment on a Post
func (handlers *FeedHandlers) CommentPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// check if custom creation timestamp is used
	var useCustomCreationTimestamp bool = false
	if createCommentRequest.CreatedAt > 0 &&
		float64(createCommentRequest.CreatedAt) <= float64(time.Now().UnixMilli()) {
		useCustomCreationTimestamp = true
	}

	// strip text and check if it is empty
	createCommentRequest.Text = strings.Trim(createCommentRequest.Text, " ")

	if createCommentRequest.Text == "" {
		utils.GeneralAPIValidationError(c, "Comment text cannot be empty")
		return
	}

	// fetch post data
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create comment using the helper method
	commentId, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, postData.ID, communityId,
		constants.CommentBaseLevel, headers[utils.HeadersMemberId], createCommentRequest.TempID, createCommentRequest.CreatedAt)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	taggedMembers := createCommentRequest.UUIDs

	var isCreatorTagged bool = false

	if !useCustomCreationTimestamp {

		// cta data for activity
		ctaData := gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  commentId.(primitive.ObjectID).Hex(),
		}

		for _, member := range taggedMembers {
			if member == postData.UserId {
				isCreatorTagged = true
			}

			// create tag activity
			activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, member,
				constants.Comment, commentId.(primitive.ObjectID), postData.UserId, constants.TaggedInPostComment, ctaData,
				false, false, primitive.NilObjectID)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			if activityID != nil {
				SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
			}

		}

		if !isCreatorTagged {
			activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
				postData.UserId, constants.Post, postData.ID, postData.UserId, constants.CommentOnPost, ctaData,
				false, false, commentId.(primitive.ObjectID))
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			if activityID != nil {
				handlers.CreateAlsoCommentedActivity(activityID, postData, headers, ctaData)
				SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
			}
		}

		// trigger comment added webhook
		err = handlers.taskDistributor.TriggerCommentAddedWebhook(commentId.(primitive.ObjectID).Hex(), headers[utils.HeadersApiKey])
		if err != nil {
			logging.Error("Error triggering comment added webhook", err)
		}

		// trigger comment tagged webhook
		err = handlers.taskDistributor.TriggerCommentTaggedWebhook(commentId.(primitive.ObjectID).Hex(), taggedMembers, headers[utils.HeadersApiKey])
		if err != nil {
			logging.Error("Error triggering comment tagged webhook", err)
		}
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, commentId.(primitive.ObjectID).Hex(), postId, commentFilterOptions, headers[utils.HeadersMemberId], false,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, false)
	if err == nil {
		response["comment"] = fetchCommentResponse
	}

	// Delete top liked comments data in post from cache
	handlers.cacheHelper.Del(fmt.Sprintf(cache.PostTopLikedCommentKey, communityId, postId))

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to edit a comment
func (handlers *FeedHandlers) EditComment(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editCommentRequest requests.EditCommentRequest
	if err := c.ShouldBindJSON(&editCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// strip text and check if it is empty
	editCommentRequest.Text = strings.Trim(editCommentRequest.Text, " ")

	if editCommentRequest.Text == "" {
		utils.GeneralAPIValidationError(c, "Comment text cannot be empty")
		return
	}

	// Check if Post_id is valid
	_, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If user is not cm and is not the comment creator
	if !editCommentRequest.UserIsCm && commentData.UserId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// comment update data
	commentUpdateData := gin.H{
		"$set": gin.H{
			"text":      editCommentRequest.Text,
			"is_edited": true,
		},
	}

	// update comment data
	err = handlers.commentHelper.UpdateCommentByIdHelper(commentData.ID, commentUpdateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Generate page filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, commentId, postId, commentFilterOptions, headers[utils.HeadersMemberId], editCommentRequest.UserIsCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
		"comment": fetchCommentResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Reply on a Comment
func (handlers *FeedHandlers) ReplyComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// check if custom creation timestamp is used
	var useCustomCreationTimestamp bool = false
	if createCommentRequest.CreatedAt > 0 &&
		float64(createCommentRequest.CreatedAt) <= float64(time.Now().UnixMilli()) {
		useCustomCreationTimestamp = true
	}

	// fetch post data
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of comment level
	if commentData.Level >= constants.CommentAllowedLevel {
		utils.GeneralAPIValidationError(c, constants.CommentAllowedErrorMessage)
		return
	}

	// create comment using the helper method
	newCommentId, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, postData.ID, communityId,
		commentData.Level+1, headers[utils.HeadersMemberId], createCommentRequest.TempID, createCommentRequest.CreatedAt)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// comment update data
	commentUpdateData := gin.H{
		"$push": gin.H{
			"replies": newCommentId.(primitive.ObjectID),
		},
	}

	// update post using the helper method
	err = handlers.commentHelper.UpdateCommentByIdHelper(commentData.ID, commentUpdateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	taggedMembers, err := getTaggedUsers(createCommentRequest.Text)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	var isCreatorTagged bool = false

	if !useCustomCreationTimestamp {

		// cta data for activity
		ctaData := gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  newCommentId.(primitive.ObjectID).Hex(),
		}

		for _, member := range taggedMembers {
			if member == commentData.UserId {
				isCreatorTagged = true
			}

			// create tag activity
			activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, member,
				constants.Comment, newCommentId.(primitive.ObjectID), postData.UserId, constants.TaggedInPostComment, ctaData,
				false, false, primitive.NilObjectID)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			if activityID != nil {
				SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
			}

		}

		if !isCreatorTagged {
			// create comment activity
			activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
				commentData.UserId, constants.Comment, commentData.ID, commentData.UserId, constants.CommentOnComment, ctaData,
				false, false, newCommentId.(primitive.ObjectID))
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			if activityID != nil {
				SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
			}

		}

		// trigger comment added webhook
		err = handlers.taskDistributor.TriggerCommentAddedWebhook(newCommentId.(primitive.ObjectID).Hex(), headers[utils.HeadersApiKey])
		if err != nil {
			logging.Error("Error triggering comment added webhook", err)
		}

		// trigger comment tagged webhook
		err = handlers.taskDistributor.TriggerCommentTaggedWebhook(newCommentId.(primitive.ObjectID).Hex(), taggedMembers, headers[utils.HeadersApiKey])
		if err != nil {
			logging.Error("Error triggering comment tagged webhook", err)
		}
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, newCommentId.(primitive.ObjectID).Hex(), postId, commentFilterOptions,
		headers[utils.HeadersMemberId], false, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check, false)
	if err == nil {
		response["comment"] = fetchCommentResponse
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Delete a Comment
func (handlers *FeedHandlers) DeleteComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deleteCommentRequest requests.DeleteCommentRequest
	if err := c.ShouldBindJSON(&deleteCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != commentData.UserId {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// comment update data
	commentUpdateData := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": deleteCommentRequest.DeleteReason,
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	// update post using the helper method
	err = handlers.commentHelper.UpdateCommentByIdHelper(commentData.ID, commentUpdateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// remove user post comment activity
	deleteUserPostCommentActivity(handlers, postData, c, headers)

	// remove other activity for the comment
	deleteActivityFilter := gin.H{
		"entity_type": constants.Comment,
		"entity_id":   commentData.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// create delete activity if deleted by CM
	if deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != commentData.UserId {
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
			commentData.UserId, constants.Comment, commentData.ID, commentData.UserId, constants.CMDeletedComment,
			gin.H{}, false, false, primitive.NilObjectID)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
		}

	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func deleteUserPostCommentActivity(handlers *FeedHandlers, postData *entities.Post, c *gin.Context, headers map[string]string) {

	activityFilterData := gin.H{
		"community_id": postData.CommunityId,
		"entity_type":  constants.Post,
		"entity_id":    postData.ID,
		"action":       constants.CommentOnPost,
	}

	activity, err := handlers.activityHelper.FindActivityHelper(activityFilterData, gin.H{})
	if err != nil {
		return
	}

	if activity == nil {
		return
	}

	actionBy := activity[0].ActionBy
	actionByMetadata := activity[0].ActionByMetadata
	if actionByMetadata == nil {
		actionByMetadata = map[string]entities.ActionByMetadata{}
	}

	// fetch all the comments on the post
	commentFilterData := gin.H{
		"user_id":    headers[utils.HeadersMemberId],
		"post_id":    postData.ID,
		"is_deleted": false,
	}

	sortOptions := gin.H{
		"$sort": gin.H{
			"created_at": -1,
		},
	}

	comments, err := handlers.commentHelper.FindCommentHelper(commentFilterData, sortOptions)
	if err != nil {
		return
	}

	// if user still has comments on the post, do not delete activity
	if len(comments) > 0 {

		// update action_by_metadata
		commentData := comments[0]

		actionByMetadata[headers[utils.HeadersMemberId]] = entities.ActionByMetadata{
			EntityId:  commentData.ID,
			CreatedAt: commentData.CreatedAt,
		}

	} else {

		// remove uuid from like action list
		actionBy = utils.RemoveAllOccurenceStringList(activity[0].ActionBy, headers[utils.HeadersMemberId])

		// remove user's data from action_by_metadata
		delete(actionByMetadata, headers[utils.HeadersMemberId])
	}

	// activity update data
	activityUpdateData := gin.H{
		"$set": gin.H{
			"action_by":          actionBy,
			"action_by_metadata": actionByMetadata,
		},
	}

	// update activity data, exisiting activity timestamp remains same to maintain order
	err = handlers.activityHelper.UpdateActivityByIDHelper(activity[0].ID, activityUpdateData, true, true)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// if action by is [], no user comments on post, mark activity as deleted
	if len(actionBy) == 0 {
		handlers.activityHelper.DeleteActivityHelper(activityFilterData)
		handlers.activityHelper.WarmupUserActivityFeedCache(activity[0].CommunityID, activity[0].ActionOn)
	}
}

// Internal method to create mongo query to get top comments based on likes
func createTopCommentsBasedOnLikesQuery(postIds []primitive.ObjectID, sortOrder int, commentsCount interface{}) []map[string]interface{} {
	commentsFilterData := []map[string]interface{}{}

	// Add match logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$match": gin.H{
			"$and": []gin.H{
				{
					"post_id": gin.H{
						"$in": postIds,
					},
				},
				{
					"replies": gin.H{
						"$eq": []interface{}{},
					},
				},
			},
		},
	})

	// Add lookup logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$lookup": gin.H{
			"from": "like",
			"let": gin.H{
				"commentId": "$_id",
			},
			"pipeline": []gin.H{
				{
					"$match": gin.H{
						"$expr": gin.H{
							"$eq": []string{"$entity_id", "$$commentId"},
						},
						"is_deleted": false,
					},
				},
			},
			"as": "likes_data",
		},
	})

	commentsFilterData = append(commentsFilterData, gin.H{
		"$project": gin.H{
			"post_id":    1,
			"comment_id": "$_id",
			"created_at": 1,
			"updated_at": 1,
			"likes_count": gin.H{
				"$size": "$likes_data",
			},
		},
	})

	// Add sort logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$sort": gin.H{
			"likes_count": sortOrder,
			"updated_at":  -1,
		},
	})

	// Add group logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$group": gin.H{
			"_id": "$post_id",
			"filtered_comments": gin.H{
				"$push": "$$ROOT",
			},
		},
	})

	// Add project logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$project": gin.H{
			"_id": 1,
			"top_comments": gin.H{
				"$slice": []interface{}{"$filtered_comments", commentsCount},
			},
		},
	})

	return commentsFilterData
}

// Internal Method to get top n comments against posts based on sorting key, sort order
func getTopCommentsAgainstPostsOnLikes(handlers *FeedHandlers, postIds []primitive.ObjectID, sortOrder int, commentsCount interface{}, communityId int) (map[string]interface{}, []string, error) {
	postsTopComments := map[string]interface{}{}
	allCommentsIds := []string{}

	commentsFilterData := createTopCommentsBasedOnLikesQuery(postIds, sortOrder, commentsCount)

	// fetch post using helper method
	commentResults, err := handlers.commentHelper.AggregateTopCommentsHelper(commentsFilterData)

	if err != nil {
		return nil, nil, err
	}

	for _, commentResult := range commentResults {
		var topCommentsList []string

		for _, topComment := range commentResult.TopComments {
			topCommentsList = append(topCommentsList, topComment.CommentID.Hex())
			allCommentsIds = append(allCommentsIds, topComment.CommentID.Hex())
		}

		postsTopComments[commentResult.PostID.Hex()] = topCommentsList
	}

	// Save data in cache
	for _, postId := range postIds {
		var likedCommentsCacheData []byte
		topCommentsData, ok := postsTopComments[postId.Hex()]

		if ok {
			likedCommentsCacheData, _ = json.Marshal(topCommentsData)
		}

		if likedCommentsCacheData == nil {
			likedCommentsCacheData, _ = json.Marshal([]string{})
		}

		handlers.cacheHelper.Set(fmt.Sprintf(cache.PostTopLikedCommentKey, communityId, postId.Hex()), likedCommentsCacheData, 0)
	}

	return postsTopComments, allCommentsIds, nil
}

// Get top liked comments against posts from cache
func getTopCommentsAgainstPostsOnLikesFromCache(handlers *FeedHandlers, postIds []primitive.ObjectID, communityId int) (map[string]interface{}, []string, bool, error) {
	postsTopComments := map[string]interface{}{}
	allCommentsIds := []string{}

	for _, postId := range postIds {
		var topCommentsList []string

		postTopComments := handlers.cacheHelper.Get(fmt.Sprintf(cache.PostTopLikedCommentKey, communityId, postId.Hex()))

		if postTopComments.Val() == "" {
			return nil, nil, false, nil
		}

		topCommentsBytes, err := postTopComments.Bytes()

		if err != nil {
			return nil, nil, false, err
		}

		json.Unmarshal(topCommentsBytes, &topCommentsList)

		postsTopComments[postId.Hex()] = topCommentsList
		allCommentsIds = append(allCommentsIds, topCommentsList...)
	}

	return postsTopComments, allCommentsIds, true, nil
}

func getTopCommentsAgainstPostsSortOnLikes(handlers *FeedHandlers, postsResponse []requests.PostResponse,
	userId string, isUserCM bool, communityId int, commentSortOrderVal int, commentCount int,
	versionCode string, platformCode string, apiRevampV1Check bool) ([]requests.PostResponse, map[string]requests.CommentWithParentResponse, error) {
	var updatedPostsWithComments []requests.PostResponse
	var postIds []primitive.ObjectID

	for _, postData := range postsResponse {
		postIds = append(postIds, postData.ID)
	}

	topCommentsAgainstPostsData, allCommentIds, allPostsFetched, err := getTopCommentsAgainstPostsOnLikesFromCache(handlers, postIds, communityId)

	if !allPostsFetched {
		topCommentsAgainstPostsData, allCommentIds, err = getTopCommentsAgainstPostsOnLikes(handlers, postIds, commentSortOrderVal, commentCount, communityId)
	}

	if err != nil {
		return nil, nil, err
	}

	for _, postData := range postsResponse {
		postDataAddress := &postData
		commentsData, _ := topCommentsAgainstPostsData[postData.ID.Hex()].([]string)

		(*postDataAddress).CommentIDs = commentsData
		updatedPostsWithComments = append(updatedPostsWithComments, *postDataAddress)
	}

	filtered_comments, _ := fetchMultipleCommentsData(handlers, allCommentIds, communityId, userId, isUserCM,
		versionCode, platformCode, apiRevampV1Check)

	return updatedPostsWithComments, filtered_comments, nil
}
