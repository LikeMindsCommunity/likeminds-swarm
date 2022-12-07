package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func fetchCommentRepliesCount(helper interfaces.CommentHelper, comment_id string, post_id string) (int64, error) {
	comment_data, err := fetchComment(helper, comment_id, post_id)
	if err != nil {
		return 0, err
	}

	reply_filter_data := gin.H{
		"_id": gin.H{
			"$in": comment_data.Replies,
		},
		"is_deleted": false,
		"post_id":    post_id,
	}

	// fetch replies count using helper method
	likes_count, err := helper.CountCommentHelper(reply_filter_data)
	if err != nil {
		return 0, err
	}

	return likes_count, nil
}

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

func parseCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comment entities.Comment) requests.CommentResponse {
	likes_count, _ := fetchEntityLikesCount(likeHelper, comment.ID.Hex(), constants.CommentEntityType)
	var response requests.CommentResponse

	response.ID = comment.ID
	response.Text = comment.Text
	response.Level = comment.Level
	response.UserId = comment.UserId
	response.LikesCount = int(likes_count)
	response.IsDeleted = comment.IsDeleted

	if comment.Level == constants.CommentBaseLevel {
		replies_count, _ := fetchCommentRepliesCount(commentHelper, comment.ID.Hex(), comment.PostId.Hex())
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

func parseMultipleCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	comments []entities.Comment) []requests.CommentResponse {
	var response []requests.CommentResponse
	for _, comment := range comments {
		response = append(response, parseCommentResponse(likeHelper, commentHelper, comment))
	}

	return response
}

func parseFetchCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	raw_comment *entities.Comment, parsed_comment requests.CommentResponse,
	replies []requests.CommentResponse) requests.FetchCommentResponse {
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
			parent_comment_response := parseCommentResponse(likeHelper, commentHelper, *comment_data)
			response.ParentComment = &parent_comment_response
		}
	}

	return response
}

func (handlers *commentHandlers) FetchComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// fetch post data
	_, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
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

	comment_filter_data := gin.H{
		"_id": gin.H{
			"$in": comment_data.Replies,
		},
		"is_deleted": false,
		"post_id":    post_id,
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

	replies_response := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_results)
	comment_response := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *comment_data)
	fetch_comment_response := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper, comment_data,
		comment_response, replies_response)

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"comment": fetch_comment_response,
	})
}

func (handlers *commentHandlers) CommentPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	post_data, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create comment using the helper method
	comment_id, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, post_data.ID,
		constants.CommentBaseLevel, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// create also_comment activity
	err = createActivity(handlers.activityHelper, constants.AlsoCommentAction, post_data.ID, constants.PostEntityType,
		post_data.ApiKey, headers[utils.HeadersMemberId], post_data.UserId, gin.H{
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
		err = createActivity(handlers.activityHelper, constants.TagAction, comment_id.(primitive.ObjectID), constants.CommentEntityType,
			post_data.ApiKey, headers[utils.HeadersMemberId], member, gin.H{
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
		err = createActivity(handlers.activityHelper, constants.CommentAction, post_data.ID, constants.PostEntityType,
			post_data.ApiKey, headers[utils.HeadersMemberId], post_data.UserId, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  comment_id.(primitive.ObjectID).Hex(),
			})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *commentHandlers) ReplyComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	post_data, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
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
	new_comment_id, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, post_data.ID,
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
		err = createActivity(handlers.activityHelper, constants.TagAction, new_comment_id.(primitive.ObjectID), constants.CommentEntityType,
			post_data.ApiKey, headers[utils.HeadersMemberId], member, gin.H{
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
		err = createActivity(handlers.activityHelper, constants.CommentAction, comment_data.ID, constants.CommentEntityType,
			post_data.ApiKey, headers[utils.HeadersMemberId], comment_data.UserId, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  new_comment_id.(primitive.ObjectID).Hex(),
			})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *commentHandlers) DeleteComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of request body
	var deleteCommentRequest requests.DeleteCommentRequest
	if err := c.ShouldBindJSON(&deleteCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	post_data, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
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
	err = createActivity(handlers.activityHelper, constants.DeleteAction, comment_data.ID, constants.CommentEntityType,
		post_data.ApiKey, headers[utils.HeadersMemberId], post_data.UserId, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

type commentHandlers struct {
	likeHelper     interfaces.LikeHelper
	commentHelper  interfaces.CommentHelper
	postHelper     interfaces.PostHelper
	activityHelper interfaces.ActivityHelper
}

func NewCommentHandlers(commentHelper interfaces.CommentHelper, likeHelper interfaces.LikeHelper,
	postHelper interfaces.PostHelper, activityHelper interfaces.ActivityHelper) *commentHandlers {
	return &commentHandlers{
		likeHelper:     likeHelper,
		commentHelper:  commentHelper,
		postHelper:     postHelper,
		activityHelper: activityHelper,
	}
}
