package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
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
func fetchEntityLikesCount(helper interfaces.LikeHelper, entity_id string, entity_type string) int {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"is_deleted":  false,
	}

	// fetch likes count using helper method
	likes_count, err := helper.CountLikeHelper(like_filter_data)
	if err != nil {
		logging.Error("Failed to fetch likes count: ", err)
		return 0
	}

	return int(likes_count)
}

// Internal method to fetch likes count for multiple entities in bulk
func fetchMultipleEntitiesLikesCount(helper interfaces.LikeHelper, entity_ids []primitive.ObjectID, entity_type string,
) map[primitive.ObjectID]int {

	counts := make(map[primitive.ObjectID]int, len(entity_ids))

	// fetch likes count using helper method
	query := []map[string]interface{}{
		{
			"$match": gin.H{
				"entity_id": bson.M{
					"$in": entity_ids,
				},
				"entity_type": entity_type,
				"is_deleted":  false,
			},
		},
		{
			"$group": gin.H{
				"_id": "$entity_id",
				"count": gin.H{
					"$sum": 1,
				},
			},
		},
	}

	results, err := helper.AggregateLikeHelper(query)
	if err != nil {
		logging.Error("Failed to fetch likes count: ", err)
		return counts
	}

	for _, result := range results.([]gin.H) {
		counts[result["_id"].(primitive.ObjectID)] = int(result["count"].(int32))
	}

	return counts
}

// Internal Method to fetch the likes for a specific Entity
func fetchEntityLikes(helper interfaces.LikeHelper, entity_id string, entity_type string,
	filterOpts map[string]interface{}, excludedUsersList []string) ([]entities.Like, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"is_deleted":  false,
		"liked_by": gin.H{
			"$nin": excludedUsersList,
		},
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
func fetchUserLikedStatusByEntity(helper interfaces.LikeHelper, entity_id string, entity_type string, liked_by string,
) bool {
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

func fetchUserLikedStatusForMultipleEntities(helper interfaces.LikeHelper, entityIds []primitive.ObjectID, entityType string, likedBy string,
) map[primitive.ObjectID]bool {

	likedStatus := make(map[primitive.ObjectID]bool, len(entityIds))

	likeFilterData := gin.H{
		"entity_id": bson.M{
			"$in": entityIds,
		},
		"entity_type": entityType,
		"liked_by":    likedBy,
	}

	like_results, err := helper.FindLikeHelper(likeFilterData, gin.H{})
	if err != nil {
		return likedStatus
	}

	for _, like := range like_results {
		likedStatus[like.EntityId] = !like.IsDeleted
	}

	return likedStatus
}

// Exposed Method to like a Post
func (handlers *FeedHandlers) LikePost(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	memberRole := headers[utils.HeadersMemberRole]
	userId := headers[utils.HeadersMemberId]
	post_id := c.Param("post_id")

	isCm := utils.IsCMRole(memberRole)

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var likePostRequest requests.LikeRequest
	c.ShouldBindJSON(&likePostRequest)

	// check if custom creation timestamp is used
	var useCustomCreationTimestamp bool = false
	if likePostRequest.CreatedAt > 0 &&
		float64(likePostRequest.CreatedAt) <= float64(time.Now().UnixMilli()) {
		useCustomCreationTimestamp = true
	}

	// fetch post using helper method
	postData, err := FetchPostData(handlers.postHelper, post_id, community_id, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		utils.GeneralAPIValidationError(c, utils.PostIsHiddenError)
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
		_, err := handlers.likeHelper.CreateLikeHelper(constants.PostEntityType, postData.ID,
			headers[utils.HeadersMemberId], likePostRequest.CreatedAt)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if !useCustomCreationTimestamp && !postData.IsHidden {
			createUserPostLikeActivity(handlers, postData, c, headers)

			// Trigger post liked webhook
			err := handlers.taskDistributor.TriggerPostLikedWebhook(post_id, headers[utils.HeadersMemberId], headers[utils.HeadersApiKey])
			if err != nil {
				logging.Error("Failed to trigger post liked webhook: ", err)
			}
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

		if !useCustomCreationTimestamp && !postData.IsHidden {
			if !like_data.IsDeleted {
				deleteUserPostLikeActivity(handlers, postData, c, headers)
			} else {
				createUserPostLikeActivity(handlers, postData, c, headers)
			}
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func createUserPostLikeActivity(handlers *FeedHandlers, postData *entities.Post, c *gin.Context, headers map[string]string) {

	ctaData := gin.H{
		"entity_type": constants.PostEntityType,
		"post_id":     postData.ID.Hex(),
	}

	// create post like activity
	activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
		postData.UserId, constants.PostEntity, postData.ID, postData.UserId, constants.LikeOnPost, ctaData,
		false, false, primitive.NilObjectID, "")
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activityID != nil {
		err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
		if err != nil {
			logging.Error("Failed to enqueue send notification : ", err)
		}
	}
}

func deleteUserPostLikeActivity(handlers *FeedHandlers, postData *entities.Post, c *gin.Context, headers map[string]string) {

	activityFilterData := gin.H{
		"community_id": postData.CommunityId,
		"entity_type":  constants.PostEntity,
		"entity_id":    postData.ID,
		"action":       constants.LikeOnPost,
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

	// remove action by metadata
	actionByMetadata := activity[0].ActionByMetadata
	delete(actionByMetadata, headers[utils.HeadersMemberId])

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
	}
}

// Exposed Method to fetch the likes on a Post
func (handlers *FeedHandlers) FetchPostLikes(c *gin.Context) {

	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	userId := headers[utils.HeadersMemberId]
	memberRole := headers[utils.HeadersMemberRole]

	isCm := utils.IsCMRole(memberRole)

	// fetch url params
	post_id := c.Param("post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, post_id, communityId, true, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		utils.GeneralAPIValidationError(c, utils.PostIsHiddenError)
		return
	}

	// fetch likes count using helper method
	likes_count := fetchEntityLikesCount(handlers.likeHelper, post_id, constants.PostEntityType)

	// filter options
	like_filter_options, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := fetchEntityLikes(handlers.likeHelper, post_id, constants.PostEntityType, like_filter_options, excludedUserIds)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, parseFetchLikeResponse(like_results, likes_count, apiRevampV1Check))
}

// Exposed Method to like a Comment
func (handlers *FeedHandlers) LikeComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var likeCommentRequest requests.LikeRequest
	c.ShouldBindJSON(&likeCommentRequest)

	// check if custom creation timestamp is used
	var useCustomCreationTimestamp bool = false
	if likeCommentRequest.CreatedAt > 0 &&
		float64(likeCommentRequest.CreatedAt) <= float64(time.Now().UnixMilli()) {
		useCustomCreationTimestamp = true
	}

	//fetch post using helper method
	post_data, err := FetchPostData(handlers.postHelper, post_id, community_id, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id, []string{})
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
			headers[utils.HeadersMemberId], likeCommentRequest.CreatedAt)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if !useCustomCreationTimestamp {
			createUserCommentLikeActivity(handlers, post_data, comment_data, c, headers)

			// Trigger comment liked webhook
			err := handlers.taskDistributor.TriggerCommentReactWebhook(comment_id, headers[utils.HeadersMemberId], headers[utils.HeadersApiKey])
			if err != nil {
				logging.Error("Failed to trigger comment liked webhook: ", err)
			}
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

		if !useCustomCreationTimestamp {
			if !like_data.IsDeleted {
				deleteUserCommentLikeActivity(handlers, post_data, comment_data, c, headers)
			} else {
				createUserCommentLikeActivity(handlers, post_data, comment_data, c, headers)
			}
		}
	}

	// Delete top liked comments data in post from cache
	handlers.cacheHelper.Del(fmt.Sprintf(cache.PostTopLikedCommentKey, community_id, post_id))

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func createUserCommentLikeActivity(handlers *FeedHandlers, postData *entities.Post, commentData *entities.Comment, c *gin.Context, headers map[string]string) {

	ctaData := gin.H{
		"entity_type": constants.CommentEntityType,
		"post_id":     postData.ID.Hex(),
		"comment_id":  commentData.ID.Hex(),
	}

	// create comment like activity
	activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
		commentData.UserId, constants.CommentEntity, commentData.ID, commentData.UserId, constants.LikeOnComment, ctaData,
		false, false, primitive.NilObjectID, "")
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if activityID != nil {
		err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
		if err != nil {
			logging.Error("Failed to enqueue send notification : ", err)
		}
	}
}

func deleteUserCommentLikeActivity(handlers *FeedHandlers, postData *entities.Post, commentData *entities.Comment, c *gin.Context,
	headers map[string]string) {

	activityFilterData := gin.H{
		"community_id": postData.CommunityId,
		"entity_type":  constants.CommentEntity,
		"entity_id":    commentData.ID,
		"action":       constants.LikeOnComment,
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

	// remove user's action by metadata
	actionByMetadata := activity[0].ActionByMetadata
	delete(actionByMetadata, headers[utils.HeadersMemberId])

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
	}

}

// Exposed Method to fetch likes on a Comment
func (handlers *FeedHandlers) FetchCommentLikes(c *gin.Context) {

	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	userId := headers[utils.HeadersMemberId]

	// fetch url params
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// fetch post data
	_, err = FetchPostData(handlers.postHelper, post_id, communityId, true, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	_, err = fetchComment(handlers.commentHelper, comment_id, post_id, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch likes count using helper method
	likes_count := fetchEntityLikesCount(handlers.likeHelper, comment_id, constants.CommentEntityType)

	// filter options
	like_filter_options, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := fetchEntityLikes(handlers.likeHelper, comment_id, constants.CommentEntityType, like_filter_options, excludedUserIds)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, parseFetchLikeResponse(like_results, likes_count, apiRevampV1Check))
}
