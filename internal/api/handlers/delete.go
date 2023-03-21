package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Exposed method to delete user data including all user posts & comments
func (handlers *FeedHandlers) DeleteUserData(c *gin.Context) {

	// fetch url params and headers
	headers := utils.GetHeaders(c)
	user_id := c.Param("user_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deleteUserDataRequest requests.DeleteUserDataRequest
	if err := c.ShouldBindJSON(&deleteUserDataRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deleteUserDataRequest.UserIsCm && headers[utils.HeadersMemberId] != user_id {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// update data
	update_data := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": "Invalid User",
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	// fetch user posts
	user_posts, err := handlers.postHelper.FindPostHelper(gin.H{"user_id": user_id, "is_deleted": false, "community_id": community_id}, gin.H{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update all user posts to deleted
	for _, post := range user_posts {
		err = handlers.postHelper.UpdatePostByIdHelper(post.ID, update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// fetch user comments
	user_comments, err := handlers.commentHelper.FindCommentHelper(gin.H{"user_id": user_id, "is_deleted": false, "community_id": community_id}, gin.H{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update all user comments to deleted
	for _, comment := range user_comments {
		err = handlers.commentHelper.UpdateCommentByIdHelper(comment.ID, update_data)
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
