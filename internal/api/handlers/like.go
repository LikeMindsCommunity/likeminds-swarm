package handlers

import (
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

// Internal Method to parse likes as response
func parseLikeResponse(like entities.Like, apiRevampV1Check bool) requests.LikeResponse {
	var response requests.LikeResponse

	response.ID = like.ID
	response.UserId = like.LikedBy
	response.UUID = like.LikedBy
	response.CreatedAt = int(like.CreatedAt.UnixMilli())
	response.UpdatedAt = int(like.UpdatedAt.UnixMilli())

	if apiRevampV1Check {
		response.UserId = ""
	}

	return response
}

// Internal Method to parse multiple likes for response
func parseMultipleLikeResponse(likes []entities.Like, apiRevampV1Check bool) []requests.LikeResponse {
	response := []requests.LikeResponse{}

	for _, like := range likes {
		response = append(response, parseLikeResponse(like, apiRevampV1Check))
	}

	return response
}

// Internal Method to parse like response for Fetch Likes API
func parseFetchLikeResponse(likes []entities.Like, total_count int, apiRevampV1Check bool) requests.FetchLikesResponse {
	var response requests.FetchLikesResponse

	response.Success = true
	response.TotalCount = total_count
	response.Likes = parseMultipleLikeResponse(likes, apiRevampV1Check)

	return response
}

// Internal Method to fetch likes count for a specific Entity
func fetchEntityLikesCount(helper interfaces.LikeHelper, entity_id string, entity_type string) (int64, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"is_deleted":  false,
	}

	// fetch likes count using helper method
	likes_count, err := helper.CountLikeHelper(like_filter_data)
	if err != nil {
		return 0, err
	}

	return likes_count, nil
}

// Internal Method to fetch the likes for a specific Entity
func fetchEntityLikes(helper interfaces.LikeHelper, entity_id string, entity_type string,
	filterOpts map[string]interface{}) ([]entities.Like, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"is_deleted":  false,
	}

	// fetch like using helper method
	like_results, err := helper.FindLikeHelper(like_filter_data, filterOpts)
	if err != nil {
		return nil, err
	}

	return like_results, nil
}

// Internal Method to fetch a specific user like on an Entity
func fetchSpecificMemberLikesOnEntity(helper interfaces.LikeHelper, entity_id string, entity_type string,
	member_id string) ([]entities.Like, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"liked_by":    member_id,
	}

	// fetch like using helper method
	like_results, err := helper.FindLikeHelper(like_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	return like_results, nil
}

// Internal Method to fetch the like status of a user on an Entity
func fetchUserLikedStatusByEntity(helper interfaces.LikeHelper, entity_id string, entity_type string, liked_by string) bool {
	like_results, err := fetchSpecificMemberLikesOnEntity(helper, entity_id, entity_type, liked_by)
	if err != nil {
		return false
	}

	if len(like_results) == 0 {
		return false
	}

	if like_results[0].IsDeleted {
		return false
	}

	return true
}

// Exposed Method to like a Post
func (handlers *FeedHandlers) LikePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch member like on entity
	like_results, err := fetchSpecificMemberLikesOnEntity(handlers.likeHelper, post_id, constants.PostEntityType,
		headers[utils.HeadersMemberId])
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

		createLikeActivity(handlers, post_data, c, headers)

	} else {
		like_data := like_results[0]

		// like update data
		like_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !like_data.IsDeleted,
			},
		}

		// update like using the helper method
		err = handlers.likeHelper.UpdateLikeByIdHelper(like_data.ID, like_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if !like_data.IsDeleted {
			deleteLikeActivity(handlers, post_data, c, headers)
		} else {
			createLikeActivity(handlers, post_data, c, headers)
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func createLikeActivity(
	handlers *FeedHandlers,
	postData *entities.Post,
	c *gin.Context,
	headers map[string]string) {
	// create like activity
	activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, postData.UserId, constants.Post, postData.ID, postData.UserId, constants.LikeOnPost, gin.H{
		"entity_type": constants.PostEntityType,
		"post_id":     postData.ID,
	}, false, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activityID != nil {
		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}
}

func deleteLikeActivity(
	handlers *FeedHandlers,
	postData *entities.Post,
	c *gin.Context,
	headers map[string]string) {

	activityFilterData := gin.H{
		"communityID": postData.CommunityId,
		"entityType":  constants.PostEntityType,
		"entityID":    postData.ID,
		"action":      constants.LikeOnPost,
	}

	activity, err := handlers.activityHelper.FindActivityHelper(activityFilterData, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activity == nil {
		return
	}

	// remove uuid from like action list
	actionBy := utils.RemoveAllOccurenceStringList(activity[0].ActionBy, headers[utils.HeadersMemberId])

	// activity update data
	activityUpdateData := gin.H{
		"$set": gin.H{
			"actionBy": actionBy,
		},
	}

	// update activity data, exisiting activity timestamp remains same to maintain order
	err = handlers.activityHelper.UpdateActivityByIDHelper(activity[0].ID, activityUpdateData, true)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}
}

// Exposed Method to fetch the likes on a Post
func (handlers *FeedHandlers) FetchPostLikes(c *gin.Context) {

	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// fetch url params
	post_id := c.Param("post_id")

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

	// fetch likes count using helper method
	likes_count, err := fetchEntityLikesCount(handlers.likeHelper, post_id, constants.PostEntityType)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options
	like_filter_options, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := fetchEntityLikes(handlers.likeHelper, post_id, constants.PostEntityType, like_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, parseFetchLikeResponse(like_results, int(likes_count), apiRevampV1Check))
}

// Exposed Method to like a Comment
func (handlers *FeedHandlers) LikeComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	//fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch member like on entity
	like_results, err := fetchSpecificMemberLikesOnEntity(handlers.likeHelper, comment_id, constants.CommentEntityType,
		headers[utils.HeadersMemberId])
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
		err = handlers.likeHelper.UpdateLikeByIdHelper(like_data.ID, like_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// create like activity
	activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, comment_data.UserId, constants.Comment, comment_data.ID, comment_data.UserId, constants.LikeOnComment, gin.H{
		"entity_type": constants.CommentEntityType,
		"post_id":     post_id,
		"comment_id":  comment_id,
	}, false, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activityID != nil {
		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// Exposed Method to fetch likes on a Comment
func (handlers *FeedHandlers) FetchCommentLikes(c *gin.Context) {

	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// fetch url params
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

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

	// fetch comment using helper method
	_, err = fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch likes count using helper method
	likes_count, err := fetchEntityLikesCount(handlers.likeHelper, comment_id, constants.CommentEntityType)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options
	like_filter_options, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := fetchEntityLikes(handlers.likeHelper, comment_id, constants.CommentEntityType, like_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, parseFetchLikeResponse(like_results, int(likes_count), apiRevampV1Check))
}
