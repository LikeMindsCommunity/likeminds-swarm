package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (handlers *commentHandlers) CommentPostOrComment(c *gin.Context) {
	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")
	comment_id = comment_id[1:]

	// post filter data
	post_filter_data := gin.H{
		"_id":        post_id,
		"is_deleted": false,
		"api_key":    headers[utils.HeadersApiKey],
	}

	// fetch post using helper method
	results, err := handlers.postHelper.FindPostHelper(post_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// validation of post_id
	if len(results) == 0 {
		utils.GeneralAPIValidationError(c, "Invalid post_id sent")
		return
	}

	post_data := results[0]

	// if comment on post
	if comment_id == "" {
		// create comment using the helper method
		_, err = handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, post_data.ID,
			constants.CommentBaseLevel, headers[utils.HeadersMemberId])
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// TODO- create activity of comment and handle tag logic

		// return final response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	} else {
		// if reply on comment

		// comment filter data
		comment_filter_data := gin.H{
			"_id":        comment_id,
			"is_deleted": false,
			"post_id":    post_id,
		}

		// fetch comment using helper method
		results, err := handlers.commentHelper.FindCommentHelper(comment_filter_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// validation of comment
		if len(results) == 0 {
			utils.GeneralAPIValidationError(c, "Invalid comment_id sent")
			return
		}

		comment_data := results[0]

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

		// update comment using the helper method
		err = handlers.commentHelper.UpdateCommentHelper(comment_filter_data, comment_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// TODO- create activity of comment and handle tag logic

		// return final response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}

func (handlers *commentHandlers) DeleteComment(c *gin.Context) {
	// validation of request body
	var deleteCommentRequest requests.DeleteCommentRequest
	if err := c.ShouldBindJSON(&deleteCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// post filter data
	post_filter_data := gin.H{
		"_id":        post_id,
		"is_deleted": false,
		"api_key":    headers[utils.HeadersApiKey],
	}

	// fetch post using helper method
	post_results, err := handlers.postHelper.FindPostHelper(post_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// validation of post_id
	if len(post_results) == 0 {
		utils.GeneralAPIValidationError(c, "Invalid post_id sent")
		return
	}

	// validation on comment_id
	if comment_id == "" {
		utils.GeneralAPIValidationError(c, "Send comment_id in url")
		return
	}

	// comment filter data
	comment_filter_data := gin.H{
		"_id":        comment_id,
		"is_deleted": false,
		"post_id":    post_id,
	}

	// fetch comment using helper method
	comment_results, err := handlers.commentHelper.FindCommentHelper(comment_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// validation of comment
	if len(comment_results) == 0 {
		utils.GeneralAPIValidationError(c, "Invalid comment_id sent")
		return
	}

	comment_data := comment_results[0]

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

	// update comment using the helper method
	err = handlers.commentHelper.UpdateCommentHelper(comment_filter_data, comment_update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// TODO- create activity of post deletion

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

type commentHandlers struct {
	commentHelper interfaces.CommentHelper
	postHelper    interfaces.PostHelper
}

func NewCommentHandlers(commentHelper interfaces.CommentHelper, postHelper interfaces.PostHelper) *commentHandlers {
	return &commentHandlers{
		commentHelper: commentHelper,
		postHelper:    postHelper,
	}
}
