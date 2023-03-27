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
	comment entities.Comment, user_id string, is_cm bool) requests.CommentResponse {
	likes_count, _ := fetchEntityLikesCount(likeHelper, comment.ID.Hex(), constants.CommentEntityType)
	var response requests.CommentResponse

	response.ID = comment.ID
	response.Text = comment.Text
	response.Level = comment.Level
	response.CommunityId = comment.CommunityId
	response.UserId = comment.UserId
	response.IsLiked = fetchUserLikedStatusByEntity(likeHelper, comment.ID.Hex(), constants.CommentEntityType, user_id)
	response.LikesCount = int(likes_count)
	response.IsDeleted = comment.IsDeleted
	response.MenuItems = parseMenuItems(getEntityMenuItems(constants.CommentEntityType, is_cm, user_id == comment.UserId, false))

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
	comments []entities.Comment, user_id string, is_cm bool) []requests.CommentResponse {
	var response []requests.CommentResponse
	for _, comment := range comments {
		response = append(response, parseCommentResponse(likeHelper, commentHelper, comment, user_id, is_cm))
	}

	return response
}

// Internal Method to parse comment response for FetchComment API
func parseFetchCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	raw_comment *entities.Comment, parsed_comment requests.CommentResponse,
	replies []requests.CommentResponse, user_id string, is_cm bool) requests.FetchCommentResponse {
	var response requests.FetchCommentResponse

	response.CommentResponse = parsed_comment

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []requests.CommentResponse{}
	}

	if parsed_comment.Level > constants.CommentBaseLevel {
		comment_data, err := fetchParentComment(commentHelper, raw_comment.ID, raw_comment.PostId)
		if err == nil {
			parent_comment_response := parseCommentResponse(likeHelper, commentHelper, *comment_data, user_id, is_cm)
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

	replies_response := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_results, headers[utils.HeadersMemberId], is_cm)
	comment_response := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *comment_data, headers[utils.HeadersMemberId], is_cm)
	fetch_comment_response := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_data,
		comment_response, replies_response, headers[utils.HeadersMemberId], is_cm)

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"comment": fetch_comment_response,
	})
}

func fetchCommentData(handlers *FeedHandlers, comment_id string, post_id string, filter_options map[string]interface{}, member_id string, is_cm bool) (interface{}, error) {
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

	replies_response := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_results, member_id, is_cm)
	comment_response := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *comment_data, member_id, is_cm)
	fetch_comment_response := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_data,
		comment_response, replies_response, member_id, is_cm)

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
	fetch_comment_response, err := fetchCommentData(handlers, comment_id, post_id, comment_filter_options, headers[utils.HeadersMemberId], is_cm)
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
	_, err = createActivity(*handlers, constants.AlsoCommentAction, post_data.ID, constants.PostEntityType,
		post_data.CommunityId, headers[utils.HeadersMemberId], post_data.UserId, gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     post_id,
		})
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
		if member == post_data.UserId {
			is_creator_tagged = true
		}

		// create tag activity
		_, err = createActivity(*handlers, constants.TagAction, comment_id.(primitive.ObjectID), constants.CommentEntityType,
			post_data.CommunityId, headers[utils.HeadersMemberId], member, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  comment_id.(primitive.ObjectID).Hex(),
			})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	if !is_creator_tagged {
		// create comment activity
		_, err = createActivity(*handlers, constants.CommentAction, post_data.ID, constants.PostEntityType,
			post_data.CommunityId, headers[utils.HeadersMemberId], post_data.UserId, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  comment_id.(primitive.ObjectID).Hex(),
			})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
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
	fetch_comment_response, err := fetchCommentData(handlers, comment_id.(primitive.ObjectID).Hex(), post_id, comment_filter_options, headers[utils.HeadersMemberId], false)
	if err == nil {
		response["comment"] = fetch_comment_response
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
		_, err = createActivity(*handlers, constants.TagAction, new_comment_id.(primitive.ObjectID), constants.CommentEntityType,
			post_data.CommunityId, headers[utils.HeadersMemberId], member, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  new_comment_id.(primitive.ObjectID).Hex(),
			})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	if !is_creator_tagged {
		// create comment activity
		_, err = createActivity(*handlers, constants.CommentAction, comment_data.ID, constants.CommentEntityType,
			post_data.CommunityId, headers[utils.HeadersMemberId], comment_data.UserId, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  new_comment_id.(primitive.ObjectID).Hex(),
			})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
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
	fetch_comment_response, err := fetchCommentData(handlers, new_comment_id.(primitive.ObjectID).Hex(), post_id, comment_filter_options, headers[utils.HeadersMemberId], false)
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

	// create delete activity
	_, err = createActivity(*handlers, constants.DeleteAction, comment_data.ID, constants.CommentEntityType,
		post_data.CommunityId, headers[utils.HeadersMemberId], post_data.UserId, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
