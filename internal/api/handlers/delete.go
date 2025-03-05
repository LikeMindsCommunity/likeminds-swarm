package handlers

import (
	"fmt"
	"net/http"

	"github.com/nateshr/likeminds-swarm/internal/helpers"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Exposed method to delete user data including all user posts & comments
func (handlers *FeedHandlers) DeleteUserData(c *gin.Context) {

	// fetch url params and headers
	headers := utils.GetHeaders(c)

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deleteUserDataRequest requests.DeleteUserDataRequest
	if err := c.ShouldBindJSON(&deleteUserDataRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Check if user is not cm if user_ids are more than 1
	if !deleteUserDataRequest.UserIsCm && len(deleteUserDataRequest.UserIds) > 1 {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Check if user is not cm and user_id is same as member_id
	if !deleteUserDataRequest.UserIsCm && deleteUserDataRequest.UserIds[0] != headers[utils.HeadersMemberId] {
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

	// Iterate over all user ids and delete all posts and comments
	for _, user_id := range deleteUserDataRequest.UserIds {

		// fetch user posts
		user_posts, err := handlers.postHelper.FindPostHelper(gin.H{"user_id": user_id, "is_deleted": false, "community_id": community_id}, gin.H{})
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		// soft delete all user posts
		for _, post := range user_posts {
			err = handlers.postHelper.UpdatePostByIdHelper(post.ID, update_data)
			if err != nil {
				log.Error(fmt.Sprintf("DeleteUserData() - Error while deleting post with _id %s : %s", post.ID.Hex(), err.Error()))
			}

			// update the count of posts in topics
			if len(post.TopicIds) > 0 {
				stringTopicIds := helpers.ParseObjectIdsToString(post.TopicIds)
				err = handlers.esHelper.UpdateByQuery(UpdatePostCountInTopicsQuery(stringTopicIds, false), constants.TopicIndexName)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}

		// fetch user comments
		user_comments, err := handlers.commentHelper.FindCommentHelper(gin.H{"user_id": user_id, "is_deleted": false, "community_id": community_id}, gin.H{})
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		// soft delete all user comments
		for _, comment := range user_comments {
			err = handlers.commentHelper.UpdateCommentByIdHelper(comment.ID, update_data)
			if err != nil {
				log.Error(fmt.Sprintf("DeleteUserData() - Error while deleting comment with _id %s : %s", comment.ID.Hex(), err.Error()))
			}
		}

	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
