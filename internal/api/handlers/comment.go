package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
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
	apiRevampV1Check bool) (map[string]requests.FetchCommentsResponse, error) {

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
	parsedCommentsResponse := map[string]requests.FetchCommentsResponse{}
	for _, comment := range comments {

		// Parse comment for response
		parseCommentsResponse := parseCommentsResponse(handlers, comment, userId, isCm, versionCode, platformCode, apiRevampV1Check)

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
	apiRevampV1Check bool) requests.CommentResponse {
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
	response.MenuItems = getEntityMenuItems(constants.CommentEntityType, isCm, userId == comment.UserId, false, versionCode, platformCode)

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

// Internal Method to parse multiple comments for response
func parseMultipleCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comments []entities.Comment, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool) []requests.CommentResponse {
	var response []requests.CommentResponse
	for _, comment := range comments {
		response = append(response, parseCommentResponse(likeHelper, commentHelper, comment, userId, isCm, versionCode, platformCode,
			apiRevampV1Check))
	}

	return response
}

// Internal method to parse comments for FetchCommentsResponse
func parseCommentsResponse(handlers *FeedHandlers, comment entities.Comment, userId string, isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool) requests.FetchCommentsResponse {

	fetchCommentResponse := requests.FetchCommentsResponse{
		CommentResponse: parseCommentResponse(handlers.likeHelper, handlers.commentHelper, comment, userId, isCm, versionCode, platformCode,
			apiRevampV1Check),
	}

	// Fetch parent comment if exists
	if fetchCommentResponse.Level > constants.CommentBaseLevel {
		parentComment, err := fetchParentComment(handlers.commentHelper, comment.ID, comment.PostId)
		if err == nil {
			parentCommentResponse := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *parentComment, userId, isCm, versionCode, platformCode,
				apiRevampV1Check)
			fetchCommentResponse.ParentComment = &parentCommentResponse
		}
	}

	return fetchCommentResponse

}

// Internal Method to parse comment data for FetchComment API
func parseFetchCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	rawComment *entities.Comment, replies []requests.CommentResponse, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool) requests.FetchCommentResponse {
	var response requests.FetchCommentResponse

	response.CommentResponse = parseCommentResponse(likeHelper, commentHelper, *rawComment, userId, isCm, versionCode, platformCode,
		apiRevampV1Check)

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []requests.CommentResponse{}
	}

	if response.CommentResponse.Level > constants.CommentBaseLevel {
		commentData, err := fetchParentComment(commentHelper, rawComment.ID, rawComment.PostId)
		if err == nil {
			parentCommentResponse := parseCommentResponse(likeHelper, commentHelper, *commentData, userId, isCm, versionCode, platformCode,
				apiRevampV1Check)
			response.ParentComment = &parentCommentResponse
		}
	}

	return response
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
	commentFilterOptions, err := generatePageFilterOptions(c, "")
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
		isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	fetchCommentResponse := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentData, repliesResponse, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check)

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

func fetchCommentData(handlers *FeedHandlers, commentId string, postId string, filterOptions map[string]interface{},
	memberId string, isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool) (interface{}, error) {
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

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, filterOptions)
	if err != nil {
		return nil, err
	}

	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, memberId, isCm,
		versionCode, platformCode, apiRevampV1Check)
	fetchCommentResponse := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentData, repliesResponse, memberId, isCm, versionCode, platformCode, apiRevampV1Check)

	return fetchCommentResponse, nil
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
	commentFilterOptions, err := generatePageFilterOptions(c, "")
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
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
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
		constants.CommentBaseLevel, headers[utils.HeadersMemberId], createCommentRequest.TempID)
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

	for _, member := range taggedMembers {
		if member == postData.UserId {
			isCreatorTagged = true
		}

		// create tag activity
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, member, constants.Comment, commentId.(primitive.ObjectID), postData.UserId, constants.TaggedInPostComment, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  commentId.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
		}

	}

	if !isCreatorTagged {
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, postData.UserId, constants.Post, postData.ID, postData.UserId, constants.CommentOnPost, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  commentId.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
		}
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "")
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
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	if err == nil {
		response["comment"] = fetchCommentResponse
	}

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
	commentFilterOptions, err := generatePageFilterOptions(c, "")
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, commentId, postId, commentFilterOptions, headers[utils.HeadersMemberId], editCommentRequest.UserIsCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
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
		commentData.Level+1, headers[utils.HeadersMemberId], createCommentRequest.TempID)
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

	for _, member := range taggedMembers {
		if member == commentData.UserId {
			isCreatorTagged = true
		}

		// create tag activity
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, member, constants.Comment, newCommentId.(primitive.ObjectID), postData.UserId, constants.TaggedInPostComment, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  newCommentId.(primitive.ObjectID).Hex(),
		}, false, false)
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
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, commentData.UserId, constants.Comment, commentData.ID, commentData.UserId, constants.CommentOnComment, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  newCommentId.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
		}

	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "")
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
		headers[utils.HeadersMemberId], false, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
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

	// remove activity for the comment
	deleteActivityFilter := gin.H{
		"entity_type": constants.Comment,
		"entity_id":   commentData.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// create delete activity if deleted if CM
	if deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != commentData.UserId {
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, commentData.UserId, constants.Comment, commentData.ID, commentData.UserId, constants.CMDeletedComment, gin.H{}, false, false)
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
