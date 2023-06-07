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
func fetchPostCommentsCount(helper interfaces.CommentHelper, post_id string) (int64, error) {
	// comment filter data
	comment_filter_data := gin.H{
		"post_id":    post_id,
		"is_deleted": false,
		"level":      constants.CommentBaseLevel,
	}

	// fetch likes count using helper method
	likes_count, err := helper.CountCommentHelper(comment_filter_data)
	if err != nil {
		return 0, err
	}

	return likes_count, nil
}

// Internal Method to fetch replies count of a Comment
func fetchCommentRepliesCount(helper interfaces.CommentHelper, comment_id string) (int64, error) {
	comment_data, err := fetchCommentByIdInternal(helper, comment_id)
	if err != nil {
		return 0, err
	}

	reply_filter_data := gin.H{
		"_id": gin.H{
			"$in": comment_data.Replies,
		},
		"is_deleted": false,
	}

	// fetch replies count using helper method
	likes_count, err := helper.CountCommentHelper(reply_filter_data)
	if err != nil {
		return 0, err
	}

	return likes_count, nil
}

// Internal Method to fetch parent comment of a Comment
func fetchParentComment(helper interfaces.CommentHelper, comment_id primitive.ObjectID,
	post_id primitive.ObjectID) (*entities.Comment, error) {
	// comment filter data
	comment_filter_data := gin.H{
		"replies":    comment_id,
		"is_deleted": false,
		"post_id":    post_id,
	}

	// fetch comment using helper method
	comment_results, err := helper.FindCommentHelper(comment_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(comment_results) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &comment_results[0], nil
}

// Internal Method to fetch a comment using comment_id
func fetchCommentByIdInternal(helper interfaces.CommentHelper, comment_id string) (*entities.Comment, error) {
	// comment filter data
	comment_filter_data := gin.H{
		"_id":        comment_id,
		"is_deleted": false,
	}

	// fetch comment using helper method
	comment_results, err := helper.FindCommentHelper(comment_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(comment_results) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &comment_results[0], nil
}

// Internal Method to fetch multiple comments data using comment_ids
func fetchMultipleCommentsData(handlers *FeedHandlers,
	comment_ids []string,
	community_id int,
	user_id string,
	is_cm bool,
	versionCode string,
	platformCode string) (map[string]requests.FetchCommentsResponse, error) {

	// convert comment_ids to object ids
	comment_object_ids := helpers.ConvertIdsToObjectIds(comment_ids)

	// comment filter data
	comment_filter_data := gin.H{
		"_id": gin.H{
			"$in": comment_object_ids,
		},
		"community_id": community_id,
	}

	// fetch comments using helper method
	comments, err := handlers.commentHelper.FindCommentHelper(comment_filter_data, nil)
	if err != nil {
		return nil, err
	}

	// Make key value pair map for response, comment_id -> comment
	parsed_comments_response := map[string]requests.FetchCommentsResponse{}
	for _, comment := range comments {

		// Parse comment for response
		parseCommentsResponse := parseCommentsResponse(handlers, comment, user_id, is_cm, versionCode, platformCode)

		// Add to response map
		parsed_comments_response[comment.ID.Hex()] = parseCommentsResponse

	}

	return parsed_comments_response, nil
}

// Internal Method to fetch a comment using comment_id and post_id
func fetchComment(helper interfaces.CommentHelper, comment_id string, post_id string) (*entities.Comment, error) {
	// comment filter data
	comment_filter_data := gin.H{
		"_id":        comment_id,
		"is_deleted": false,
		"post_id":    post_id,
	}

	// fetch comment using helper method
	comment_results, err := helper.FindCommentHelper(comment_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(comment_results) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &comment_results[0], nil
}

// Internal Method to parse comment for response
func parseCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comment entities.Comment, user_id string, is_cm bool, versionCode string, platformCode string) requests.CommentResponse {
	likes_count, _ := fetchEntityLikesCount(likeHelper, comment.ID.Hex(), constants.CommentEntityType)
	var response requests.CommentResponse

	response.ID = comment.ID
	response.Text = comment.Text
	response.Level = comment.Level
	response.CommunityId = comment.CommunityId
	response.PostId = comment.PostId
	response.UserId = comment.UserId
	response.IsLiked = fetchUserLikedStatusByEntity(likeHelper, comment.ID.Hex(), constants.CommentEntityType, user_id)
	response.LikesCount = int(likes_count)
	response.IsDeleted = comment.IsDeleted
	response.IsEdited = comment.IsEdited
	response.MenuItems = getEntityMenuItems(constants.CommentEntityType, is_cm, user_id == comment.UserId, false, versionCode, platformCode)

	if comment.Level == constants.CommentBaseLevel {
		replies_count, _ := fetchCommentRepliesCount(commentHelper, comment.ID.Hex())
		response.CommentsCount = int(replies_count)
	}

	if comment.IsDeleted {
		response.DeleteReason = comment.DeleteReason
		response.DeletedBy = comment.DeletedBy
	}

	response.CreatedAt = int(comment.CreatedAt.UnixMilli())
	response.UpdatedAt = int(comment.UpdatedAt.UnixMilli())

	return response
}

// Internal Method to parse multiple comments for response
func parseMultipleCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comments []entities.Comment, user_id string, is_cm bool,
	versionCode string, platformCode string) []requests.CommentResponse {
	var response []requests.CommentResponse
	for _, comment := range comments {
		response = append(response, parseCommentResponse(likeHelper, commentHelper, comment, user_id, is_cm, versionCode, platformCode))
	}

	return response
}

// Internal method to parse comments for FetchCommentsResponse
func parseCommentsResponse(handlers *FeedHandlers, comment entities.Comment, user_id string, is_cm bool, versionCode string, platformCode string) requests.FetchCommentsResponse {

	fetch_comment_response := requests.FetchCommentsResponse{
		CommentResponse: parseCommentResponse(handlers.likeHelper, handlers.commentHelper, comment, user_id, is_cm, versionCode, platformCode),
	}

	// Fetch parent comment if exists
	if fetch_comment_response.Level > constants.CommentBaseLevel {
		parent_comment, err := fetchParentComment(handlers.commentHelper, comment.ID, comment.PostId)
		if err == nil {
			parent_comment_response := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *parent_comment, user_id, is_cm, versionCode, platformCode)
			fetch_comment_response.ParentComment = &parent_comment_response
		}
	}

	return fetch_comment_response

}

// Internal Method to parse comment data for FetchComment API
func parseFetchCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	raw_comment *entities.Comment, replies []requests.CommentResponse, user_id string, is_cm bool,
	versionCode string, platformCode string) requests.FetchCommentResponse {
	var response requests.FetchCommentResponse

	response.CommentResponse = parseCommentResponse(likeHelper, commentHelper, *raw_comment, user_id, is_cm, versionCode, platformCode)

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []requests.CommentResponse{}
	}

	if response.CommentResponse.Level > constants.CommentBaseLevel {
		comment_data, err := fetchParentComment(commentHelper, raw_comment.ID, raw_comment.PostId)
		if err == nil {
			parent_comment_response := parseCommentResponse(likeHelper, commentHelper, *comment_data, user_id, is_cm, versionCode, platformCode)
			response.ParentComment = &parent_comment_response
		}
	}

	return response
}

// Exposed Method to fetch comment by comment_id
func (handlers *FeedHandlers) FetchCommentById(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	comment_id := c.Param("comment_id")
	param_is_cm := c.Query("user_is_cm")
	is_cm := false

	if param_is_cm == "true" {
		is_cm = true
	}

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch comment data
	comment_data, err := fetchCommentByIdInternal(handlers.commentHelper, comment_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	comment_filter_data := gin.H{
		"_id": gin.H{
			"$in": comment_data.Replies,
		},
		"is_deleted": false,
		"post_id":    comment_data.PostId.Hex(),
	}

	// filter options
	comment_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	comment_results, err := handlers.commentHelper.FindCommentHelper(comment_filter_data, comment_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	replies_response := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_results, headers[utils.HeadersMemberId],
		is_cm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	fetch_comment_response := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		comment_data, replies_response, headers[utils.HeadersMemberId], is_cm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode])

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"comment": fetch_comment_response,
	})
}

// Exposed Method to fetch multiple comments by comment_ids
func (handlers *FeedHandlers) FetchComments(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// Get Query Params
	param_comment_ids := c.Query("comment_ids")
	param_is_cm, _ := strconv.ParseBool(c.Query("user_is_cm"))

	// Check if user is CM or not
	if !param_is_cm {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Unmarshal comment_ids
	var comment_ids []string
	err := json.Unmarshal([]byte(param_comment_ids), &comment_ids)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Fetch comments using comment_ids
	comments, err := fetchMultipleCommentsData(handlers, comment_ids, community_id, headers[utils.HeadersMemberId], param_is_cm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
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

func fetchCommentData(handlers *FeedHandlers, comment_id string, post_id string, filter_options map[string]interface{}, member_id string, is_cm bool, versionCode string, platformCode string) (interface{}, error) {
	// fetch comment data
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		return nil, err
	}

	comment_filter_data := gin.H{
		"_id": gin.H{
			"$in": comment_data.Replies,
		},
		"is_deleted": false,
		"post_id":    post_id,
	}

	// fetch comment using helper method
	comment_results, err := handlers.commentHelper.FindCommentHelper(comment_filter_data, filter_options)
	if err != nil {
		return nil, err
	}

	replies_response := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_results, member_id, is_cm, versionCode, platformCode)
	fetch_comment_response := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		comment_data, replies_response, member_id, is_cm, versionCode, platformCode)

	return fetch_comment_response, nil
}

// Exposed method to fetch comment by comment_id and post_id
func (handlers *FeedHandlers) FetchComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")
	param_is_cm := c.Query("user_is_cm")
	is_cm := false

	if param_is_cm == "true" {
		is_cm = true
	}

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post data
	_, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// filter options
	comment_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch comment response data
	fetch_comment_response, err := fetchCommentData(handlers, comment_id, post_id, comment_filter_options, headers[utils.HeadersMemberId], is_cm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response["comment"] = fetch_comment_response

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to comment on a Post
func (handlers *FeedHandlers) CommentPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
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
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create comment using the helper method
	comment_id, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, post_data.ID, community_id,
		constants.CommentBaseLevel, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// create also_comment activity
	activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, post_data.UserId, constants.Post, post_data.ID, headers[utils.HeadersMemberId], constants.AlsoCommentOnPost, gin.H{
		"entity_type": constants.PostEntityType,
		"post_id":     post_id,
	}, false, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])

	tagged_members, err := getTaggedUsers(createCommentRequest.Text)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	var is_creator_tagged bool = false

	for _, member := range tagged_members {
		if member == post_data.UserId {
			is_creator_tagged = true
		}

		// create tag activity
		activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, member, constants.Comment, comment_id.(primitive.ObjectID), post_data.UserId, constants.TaggedInPostComment, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     post_id,
			"comment_id":  comment_id.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	if !is_creator_tagged {
		activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, post_data.UserId, constants.Post, post_data.ID, post_data.UserId, constants.CommentOnPost, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     post_id,
			"comment_id":  comment_id.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	// filter options
	comment_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch comment response data
	fetch_comment_response, err := fetchCommentData(handlers, comment_id.(primitive.ObjectID).Hex(), post_id, comment_filter_options, headers[utils.HeadersMemberId], false,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err == nil {
		response["comment"] = fetch_comment_response
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to edit a comment
func (handlers *FeedHandlers) EditComment(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
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
	_, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If user is not cm and is not the comment creator
	if !editCommentRequest.UserIsCm && comment_data.UserId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// comment update data
	comment_update_data := gin.H{
		"$set": gin.H{
			"text":      editCommentRequest.Text,
			"is_edited": true,
		},
	}

	// update comment data
	err = handlers.commentHelper.UpdateCommentByIdHelper(comment_data.ID, comment_update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Generate page filter options
	comment_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment response data
	fetch_comment_response, err := fetchCommentData(handlers, comment_id, post_id, comment_filter_options, headers[utils.HeadersMemberId], editCommentRequest.UserIsCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
		"comment": fetch_comment_response,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Reply on a Comment
func (handlers *FeedHandlers) ReplyComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of comment level
	if comment_data.Level >= constants.CommentAllowedLevel {
		utils.GeneralAPIValidationError(c, constants.CommentAllowedErrorMessage)
		return
	}

	// create comment using the helper method
	new_comment_id, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, post_data.ID, community_id,
		comment_data.Level+1, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// comment update data
	comment_update_data := gin.H{
		"$push": gin.H{
			"replies": new_comment_id.(primitive.ObjectID),
		},
	}

	// update post using the helper method
	err = handlers.commentHelper.UpdateCommentByIdHelper(comment_data.ID, comment_update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	tagged_members, err := getTaggedUsers(createCommentRequest.Text)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	var is_creator_tagged bool = false

	for _, member := range tagged_members {
		if member == comment_data.UserId {
			is_creator_tagged = true
		}

		// create tag activity
		activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, member, constants.Comment, new_comment_id.(primitive.ObjectID), post_data.UserId, constants.TaggedInPostComment, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     post_id,
			"comment_id":  new_comment_id.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	if !is_creator_tagged {
		// create comment activity
		activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, comment_data.UserId, constants.Comment, comment_data.ID, comment_data.UserId, constants.CommentOnComment, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     post_id,
			"comment_id":  new_comment_id.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	// filter options
	comment_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch comment response data
	fetch_comment_response, err := fetchCommentData(handlers, new_comment_id.(primitive.ObjectID).Hex(), post_id, comment_filter_options, headers[utils.HeadersMemberId], false,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err == nil {
		response["comment"] = fetch_comment_response
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Delete a Comment
func (handlers *FeedHandlers) DeleteComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deleteCommentRequest requests.DeleteCommentRequest
	if err := c.ShouldBindJSON(&deleteCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != comment_data.UserId {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// comment update data
	comment_update_data := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": deleteCommentRequest.DeleteReason,
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	// update post using the helper method
	err = handlers.commentHelper.UpdateCommentByIdHelper(comment_data.ID, comment_update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// remove activity for the comment
	deleteActivityFilter := gin.H{
		"entity_type": constants.Comment,
		"entity_id":   comment_data.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// create delete activity if deleted if CM
	if deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != comment_data.UserId {
		activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, comment_data.UserId, constants.Comment, comment_data.ID, comment_data.UserId, constants.CMDeletedComment, gin.H{}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
