package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func (handlers *likeHandlers) FetchPostLikes(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post data
	_, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// like filter data
	like_filter_data := gin.H{
		"entity_id":   post_id,
		"entity_type": constants.PostEntityType,
		"is_deleted":  false,
	}

	// fetch likes count using helper method
	likes_count, err := handlers.likeHelper.CountLikeHelper(like_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := handlers.likeHelper.FindLikeHelper(like_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"total_count": likes_count,
		"likes":       like_results,
	})
}

func (handlers *likeHandlers) LikePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post using helper method
	post_data, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// like filter data
	like_filter_data := gin.H{
		"entity_id":   post_data.ID,
		"entity_type": constants.PostEntityType,
		"liked_by":    headers[utils.HeadersMemberId],
	}

	// fetch like using helper method
	like_results, err := handlers.likeHelper.FindLikeHelper(like_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(like_results) == 0 {
		// create like using the helper method
		_, err := handlers.likeHelper.CreateLikeHelper(constants.PostEntityType, post_data.ID,
			headers[utils.HeadersMemberId])
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	} else {
		like_data := like_results[0]

		// like update data
		like_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !like_data.IsDeleted,
			},
		}

		// update like using the helper method
		err = handlers.likeHelper.UpdateLikeHelper(like_filter_data, like_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// activity filter data
	activity_filter_data := gin.H{
		"entity_id":   post_id,
		"entity_type": constants.PostEntityType,
		"action":      constants.LikeAction,
		"action_by":   headers[utils.HeadersMemberId],
		"action_on":   post_data.UserId,
	}

	// fetch activity using helper method
	activity_results, err := handlers.activityHelper.FindActivityHelper(activity_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// checking of existing like activity
	if len(activity_results) == 0 {
		// create activity using the helper method
		_, err = handlers.activityHelper.CreateActivityHelper(headers[utils.HeadersMemberId], post_data.UserId, post_data.ApiKey,
			constants.PostEntityType, post_data.ID, constants.LikeAction, gin.H{
				"entity_type": constants.PostEntityType,
				"post_id":     post_id,
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

func (handlers *likeHandlers) LikeComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// fetch post using helper method
	post_data, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// comment filter data
	comment_filter_data := gin.H{
		"_id":        comment_id,
		"is_deleted": false,
		"post_id":    post_data.ID,
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

	// like filter data
	like_filter_data := gin.H{
		"entity_id":   comment_id,
		"entity_type": constants.CommentEntityType,
		"liked_by":    headers[utils.HeadersMemberId],
	}

	// fetch like using helper method
	like_results, err := handlers.likeHelper.FindLikeHelper(like_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(like_results) == 0 {
		// create like using the helper method
		_, err := handlers.likeHelper.CreateLikeHelper(constants.CommentEntityType, comment_data.ID,
			headers[utils.HeadersMemberId])
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	} else {
		like_data := like_results[0]

		// like update data
		like_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !like_data.IsDeleted,
			},
		}

		// update like using the helper method
		err = handlers.likeHelper.UpdateLikeHelper(like_filter_data, like_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// activity filter data
	activity_filter_data := gin.H{
		"entity_id":   comment_id,
		"entity_type": constants.CommentEntityType,
		"action":      constants.LikeAction,
		"action_by":   headers[utils.HeadersMemberId],
		"action_on":   comment_data.UserId,
	}

	// fetch activity using helper method
	activity_results, err := handlers.activityHelper.FindActivityHelper(activity_filter_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// checking of existing like activity
	if len(activity_results) == 0 {
		// create activity using the helper method
		_, err = handlers.activityHelper.CreateActivityHelper(headers[utils.HeadersMemberId], comment_data.UserId, post_data.ApiKey,
			constants.CommentEntityType, comment_data.ID, constants.LikeAction, gin.H{
				"entity_type": constants.CommentEntityType,
				"post_id":     post_id,
				"comment_id":  comment_id,
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

type likeHandlers struct {
	likeHelper     interfaces.LikeHelper
	commentHelper  interfaces.CommentHelper
	postHelper     interfaces.PostHelper
	activityHelper interfaces.ActivityHelper
}

func NewLikeHandlers(postHelper interfaces.PostHelper, likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	activityHelper interfaces.ActivityHelper) *likeHandlers {
	return &likeHandlers{
		likeHelper:     likeHelper,
		commentHelper:  commentHelper,
		postHelper:     postHelper,
		activityHelper: activityHelper,
	}
}
