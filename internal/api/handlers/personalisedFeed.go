package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Method to recompute the personalised feed
func (handlers *FeedHandlers) RecomputePersonalisedFeed(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	apiKey := headers[utils.HeadersApiKey]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Compute recency metric and save it in cache
	RecencyMetricComputation(handlers, userId, communityId)

	// Compute post likes metric and save it in cache
	go PostLikesMetricComputation(handlers, userId, communityId)

	// Compute post comments metric and save it in cache
	go PostCommentsMetricComputation(handlers, userId, communityId)

	// Compute user groups metric and save it in cache
	go UserGroupsMetricComputation(handlers, userId, communityId, apiKey)

	// Compute user topics metric and save it in cache
	go UserTopicsMetricComputation(handlers, userId, communityId)

	// Compute user connection metric and save it in cache
	go UserConnectionMetricComputation(handlers, userId, communityId)

	utils.GenerateSuccessResponse(c, gin.H{})
}

// Recompute & save post ranking on the basis of recency metrics
func RecencyMetricComputation(handlers *FeedHandlers, userId string, communityId int,
) map[string]float64 {
	postsMetricMap := map[string]float64{}

	// Get data from cache
	cacheKey := fmt.Sprintf(cache.PostsRecencyMetricsKey, communityId)
	postsMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey)
	if postsMetricMapCacheValue.Val() != "" && postsMetricMapCacheValue.Val() != "null" {
		return postsMetricMap
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric when fetching weights: ", err)
		return postsMetricMap
	}

	// Filter for all posts of community
	allPostsOfCommunityFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"community_id": communityId,
				"is_deleted":   false,
			},
		},
		gin.H{
			"$project": gin.H{
				"_id":        1,
				"created_at": 1,
			},
		},
	}

	allPostsData, err := handlers.postHelper.AggregatePostHelper(allPostsOfCommunityFilter)
	if err != nil {
		logging.Error("Error in fetching all posts of community: ", communityId, " err: ", err)
		return postsMetricMap
	}

	currentTimeInSeconds := float64(time.Now().Unix())

	for _, postData := range allPostsData {
		var metricScore float64

		postCreatedAtInSeconds := float64(postData["created_at"].(primitive.DateTime).Time().Unix())
		if personalisedFeedWeights.RecencyMetrics.MaxThreshold-(currentTimeInSeconds-postCreatedAtInSeconds) > 0 {
			metricScore = computeRecencyMetricScore(
				postCreatedAtInSeconds,
				personalisedFeedWeights.RecencyMetrics.MaxThreshold,
				personalisedFeedWeights.RecencyMetrics.Weight,
				currentTimeInSeconds)
		}

		postsMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(postsMetricMap)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.CommunityMetricCacheTTLInMins*time.Minute)

	if setStatus.Err() != nil {
		logging.Error("Error in saving recency metric score in cache", setStatus.Err())
	}

	return postsMetricMap
}

// Compute recency metric score
func computeRecencyMetricScore(postCreatedAt float64, recencyMetricMaxThreshold float64, recencyMetricWeight float64,
	currentTime float64) float64 {
	return ((recencyMetricMaxThreshold - (currentTime - postCreatedAt)) / recencyMetricMaxThreshold) * recencyMetricWeight
}

// Recompute & save post ranking on the basis of likes metrics
func PostLikesMetricComputation(handlers *FeedHandlers, userId string, communityId int,
) map[string]float64 {
	postsLikesMetricMap := map[string]float64{}

	// Get data from cache
	cacheKey := fmt.Sprintf(cache.PostsLikesMetricsKey, communityId)
	postsLikesMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey) // Can use GetWithKeyExists to check if key exists
	if postsLikesMetricMapCacheValue.Val() != "" && postsLikesMetricMapCacheValue.Val() != "null" {
		return postsLikesMetricMap
	}

	// Start of computation of post likes metric
	var postRecencyMetricsMap map[string]float64
	postIdsArray := []string{}

	// Get data from cache
	postsMetricMapCacheKey := fmt.Sprintf(cache.PostsRecencyMetricsKey, communityId)
	postsMetricMapCacheValue := handlers.cacheHelper.Get(postsMetricMapCacheKey)
	if postsMetricMapCacheValue.Val() == "" || postsMetricMapCacheValue.Val() == "null" {
		return postsLikesMetricMap
	}

	err := json.Unmarshal([]byte(postsMetricMapCacheValue.Val()), &postRecencyMetricsMap)
	if err != nil {
		logging.Error("Error in unmarshalling recency metric score from cache", err)
	}

	for postId := range postRecencyMetricsMap {
		postIdsArray = append(postIdsArray, postId)
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return postsLikesMetricMap
	}

	// Filter for all likes count of posts in a community
	allPostsLikesCountOfCommunityFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"entity_type": constants.PostEntityType,
				"entity_id": gin.H{
					"$in": helpers.ConvertIdsToObjectIds(postIdsArray),
				},
				"is_deleted": false,
			},
		},
		gin.H{
			"$group": gin.H{
				"_id": "$entity_id",
				"count": gin.H{
					"$sum": 1,
				},
			},
		},
	}

	allPostsLikesData, err := handlers.likeHelper.AggregateLikeHelper(allPostsLikesCountOfCommunityFilter)
	if err != nil {
		logging.Error("Error in fetching all posts likes count of community: ", communityId, " err: ", err)
		return postsLikesMetricMap
	}

	for _, postData := range allPostsLikesData.([]gin.H) {
		var metricScore float64

		postLikesCount := float64(postData["count"].(int32))

		metricScore = computePostLikesMetricScore(
			postLikesCount,
			personalisedFeedWeights.LikesMetrics.MaxThreshold,
			personalisedFeedWeights.LikesMetrics.Weight)

		postsLikesMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(postsLikesMetricMap)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.CommunityMetricCacheTTLInMins*time.Minute)

	if setStatus.Err() != nil {
		logging.Error("Error in saving post likes metric score in cache", setStatus.Err())
	}

	return postsLikesMetricMap
}

// Compute post likes metric score
func computePostLikesMetricScore(postLikesCount float64, likesMetricMaxThreshold float64, likesMetricWeight float64) float64 {
	return utils.GetMinimumFromArray(postLikesCount, likesMetricMaxThreshold) * likesMetricWeight
}

// Recompute & save post ranking on the basis of comments metrics
func PostCommentsMetricComputation(handlers *FeedHandlers, userId string, communityId int,
) map[string]float64 {
	postsCommentsMetricMap := map[string]float64{}

	// Get data from cache
	cacheKey := fmt.Sprintf(cache.PostsCommentsMetricsKey, communityId)
	postsCommentsMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey)
	if postsCommentsMetricMapCacheValue.Val() != "" && postsCommentsMetricMapCacheValue.Val() != "null" {
		return postsCommentsMetricMap
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in fetching personalised feed weights from community configurations: ", err)
		return postsCommentsMetricMap
	}

	// Filter for all comments count of posts in a community
	allPostsCommentsCountOfCommunityFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"community_id": communityId,
				"is_deleted":   false,
				"level":        0,
			},
		},
		gin.H{
			"$group": gin.H{
				"_id": "$post_id",
				"count": gin.H{
					"$sum": 1,
				},
			},
		},
	}

	allPostsLikesData, err := handlers.commentHelper.AggregateCommentHelper(allPostsCommentsCountOfCommunityFilter)
	if err != nil {
		logging.Error("Error in fetching all posts comments count of community: ", communityId, " err: ", err)
		return postsCommentsMetricMap
	}

	for _, postData := range allPostsLikesData.([]gin.H) {
		var metricScore float64

		postCommentsCount := float64(postData["count"].(int32))

		metricScore = computePostCommentsMetricScore(
			postCommentsCount,
			personalisedFeedWeights.CommentsMetrics.MaxThreshold,
			personalisedFeedWeights.CommentsMetrics.Weight)

		postsCommentsMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(postsCommentsMetricMap)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.CommunityMetricCacheTTLInMins*time.Minute)

	if setStatus.Err() != nil {
		logging.Error("Error in saving post comments metric score in cache", setStatus.Err())
	}

	return postsCommentsMetricMap
}

// Compute post comments metric score
func computePostCommentsMetricScore(postCommentsCount float64, commentsMetricMaxThreshold float64, commentsMetricWeight float64) float64 {
	return utils.GetMinimumFromArray(postCommentsCount, commentsMetricMaxThreshold) * commentsMetricWeight
}

// Recompute & save post ranking on the basis of user group metrics
func UserGroupsMetricComputation(handlers *FeedHandlers, userId string, communityId int, apiKey string) {
	userGroupsMetricMap := map[string]float64{}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return
	}

	userFollowedChannels, err := externalHelpers.FetchUserCommunityChannels(handlers.cacheHelper, userId, communityId, apiKey)
	if err != nil {
		logging.Error("Unable to fetch user followed channels:", err)
	}

	if len(userFollowedChannels) == 0 {
		logging.Info("No user followed channels found for user: ", userId)
		return
	}

	// Filter for all posts of community
	allPostsOfCommunityChannelsFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"community_id": communityId,
				"is_deleted":   false,
				"chatroom_id": gin.H{
					"$in": userFollowedChannels,
				},
			},
		},
		gin.H{
			"$project": gin.H{
				"_id": 1,
			},
		},
	}

	allUserFollowedChannelPostsData, err := handlers.postHelper.AggregatePostHelper(allPostsOfCommunityChannelsFilter)
	if err != nil {
		logging.Error("Error in fetching all posts of community: ", communityId, " err: ", err)
		return
	}

	if len(allUserFollowedChannelPostsData) == 0 {
		logging.Error("No user followed channels post data found in db")
		return
	}

	for _, postData := range allUserFollowedChannelPostsData {

		metricScore := computeUserGroupsMetricScore(
			personalisedFeedWeights.UserGroupsMetrics.MaxThreshold,
			personalisedFeedWeights.UserGroupsMetrics.Weight)

		userGroupsMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(userGroupsMetricMap)

	cacheKey := fmt.Sprintf(cache.UserGroupsMetricsKey, communityId, userId)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.UserMetricCacheTTLInHours*time.Hour)

	if setStatus.Err() != nil {
		logging.Error("Error in saving user groups metric score in cache", setStatus.Err())
	}
}

// Compute user groups metric score
func computeUserGroupsMetricScore(userGroupsMetricMaxThreshold float64, userGroupsMetricWeight float64) float64 {
	return userGroupsMetricMaxThreshold * userGroupsMetricWeight
}

// Recompute & save post ranking on the basis of user topics metrics
func UserTopicsMetricComputation(handlers *FeedHandlers, userId string, communityId int) {
	userTopicsMetricMap := map[string]float64{}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return
	}

	// Filter for all posts of community
	allUserTopicsPostsOfCommunityFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"community_id": communityId,
				"user_id":      userId,
			},
		},
		gin.H{
			"$lookup": gin.H{
				"from":         "postTopics",
				"localField":   "topic_id",
				"foreignField": "topic_id",
				"as":           "result",
			},
		},
		gin.H{
			"$unwind": gin.H{
				"path": "$result",
			},
		},
		gin.H{
			"$project": gin.H{
				"_id":      0,
				"post_id":  "$result.post_id",
				"topic_id": 1,
			},
		},
		gin.H{
			"$group": gin.H{
				"_id": "$post_id",
				"topics_count": gin.H{
					"$sum": 1,
				},
			},
		},
	}

	allUserFollowedChannelPostsData, err := handlers.userTopicsHelper.AggregateUserTopicsHelper(allUserTopicsPostsOfCommunityFilter)
	if err != nil {
		logging.Error("Error in fetching all posts of community: ", communityId, " err: ", err)
		return
	}

	if len(allUserFollowedChannelPostsData) == 0 {
		logging.Error("No user followed topic post data found in db")
		return
	}

	for _, postData := range allUserFollowedChannelPostsData {
		var metricScore float64

		topicsCount := float64(postData["topics_count"].(int32))

		metricScore = computeUserTopicsMetricScore(
			topicsCount,
			personalisedFeedWeights.UserTopicsMetrics.MaxThreshold,
			personalisedFeedWeights.UserTopicsMetrics.Weight)

		userTopicsMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(userTopicsMetricMap)

	cacheKey := fmt.Sprintf(cache.UserTopicsMetricsKey, communityId, userId)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.UserMetricCacheTTLInHours*time.Hour)

	if setStatus.Err() != nil {
		logging.Error("Error in saving user groups metric score in cache", setStatus.Err())
	}
}

// Compute user topics metric score
func computeUserTopicsMetricScore(topicsCount float64, userTopicsMetricMaxThreshold float64, userTopicsMetricWeight float64) float64 {
	return utils.GetMinimumFromArray(topicsCount, userTopicsMetricMaxThreshold) * userTopicsMetricWeight
}

// Recompute & save post ranking on the basis of user connection metrics
func UserConnectionMetricComputation(handlers *FeedHandlers, userId string, communityId int) {
	// Check whether user connection setting is enabled or not
	if !externalHelpers.IsUserConnectionSettingEnabled(handlers.cacheHelper, userId, communityId) {
		logging.Error(fmt.Sprintf("User connection setting is disabled for community id %d: ", communityId))
		return
	}

	userConnectionMetricMap := map[string]float64{}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return
	}

	// Warm up the connection list
	WarmUpConnectionList(handlers, userId, communityId, enums.OneWayConnection)

	// Get user's connected userIds list
	userIdsMap, isDataExists := getUserConnectionDataFromCache(handlers, userId, communityId)
	if !isDataExists {
		logging.Error(fmt.Sprintf("User %s connections not exists in cache: ", userId), err)
		return
	}

	userIdsList := []string{}

	for userId, _ := range userIdsMap {
		userIdsList = append(userIdsList, userId)
	}

	// Filter for all posts of community
	allPostsOfConnectedUsersFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"community_id": communityId,
				"is_deleted":   false,
				"user_id": gin.H{
					"$in": userIdsList,
				},
			},
		},
		gin.H{
			"$project": gin.H{
				"_id": 1,
			},
		},
	}

	allConnectedUsersPostsData, err := handlers.postHelper.AggregatePostHelper(allPostsOfConnectedUsersFilter)
	if err != nil {
		logging.Error("Error in fetching all posts of community: ", communityId, " err: ", err)
		return
	}

	if len(allConnectedUsersPostsData) == 0 {
		logging.Error("No connected users post data found in db")
		return
	}

	for _, postData := range allConnectedUsersPostsData {
		metricScore := computeUserConnectionMetricScore(
			personalisedFeedWeights.UserConnectionMetrics.MaxThreshold,
			personalisedFeedWeights.UserConnectionMetrics.Weight)

		userConnectionMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(userConnectionMetricMap)

	cacheKey := fmt.Sprintf(cache.UserConnectionMetricsKey, communityId, userId)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.UserMetricCacheTTLInHours*time.Hour)

	if setStatus.Err() != nil {
		logging.Error("Error in saving user connection metric score in cache", setStatus.Err())
	}
}

// Compute user connection metric score
func computeUserConnectionMetricScore(userConnectionMetricMaxThreshold float64, userConnectionMetricWeight float64) float64 {
	return userConnectionMetricMaxThreshold * userConnectionMetricWeight
}

// Compute post dampening metrics for user
func UserDampenedMetricsComputation(handlers *FeedHandlers, userId string, communityId int) (map[string]float64, error) {
	userPostDampeningMetricMap := map[string]float64{}

	// Last 24hours unix timestamp
	dampenedPostsTimestampFrom := time.Now().Add(time.Duration(-24) * time.Hour).Unix()

	// Filter for all user dampened posts
	allDampenedPostsOfUserFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"entity_type":  enums.EntityTypePost,
				"community_id": communityId,
				"user_id":      userId,
				"epoch_timestamp": gin.H{
					"$gte": dampenedPostsTimestampFrom,
				},
			},
		},
		gin.H{
			"$group": gin.H{
				"_id": "",
				"post_ids": gin.H{
					"$addToSet": "$entity_id",
				},
			},
		},
		gin.H{
			"$unset": "_id",
		},
	}

	allDampenedPostsData, err := handlers.userEntityTimestampHelper.AggregateUserEntityTimestampHelper(allDampenedPostsOfUserFilter)
	if err != nil {
		logging.Error("Error in fetching all dampened posts of user:, ", userId, " in community: ", communityId, " err: ", err)
		return nil, err
	}

	if len(allDampenedPostsData) == 0 {
		logging.Error("No dampened posts found in db!")
		return userPostDampeningMetricMap, nil
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return userPostDampeningMetricMap, err
	}

	dampenedPostIdsList := allDampenedPostsData[0]["post_ids"].(primitive.A)

	for _, postId := range dampenedPostIdsList {
		postIdObject := postId.(primitive.ObjectID)
		metricScore := computeUserPostDampeningMetricScore(
			personalisedFeedWeights.PostDampeningMetrics.MaxThreshold,
			personalisedFeedWeights.PostDampeningMetrics.Weight)

		userPostDampeningMetricMap[postIdObject.Hex()] = metricScore
	}

	return userPostDampeningMetricMap, nil
}

// Compute user post dampening metric score
func computeUserPostDampeningMetricScore(userTopicsMetricMaxThreshold float64, userTopicsMetricWeight float64) float64 {
	return userTopicsMetricMaxThreshold * userTopicsMetricWeight
}

// Internal method to save dampened posts for user in DB
func saveDampenedPostsForUserInDb(handlers *FeedHandlers, userId string, communityId int, postIds []string) {
	// Add the posts with updated timestamp to map
	currentEpochTime := time.Now().Unix()
	handlers.userEntityTimestampHelper.CreateUserEntityTimestampHelper(userId, communityId, enums.EntityTypePost,
		helpers.ConvertIdsToObjectIds(postIds), int(currentEpochTime))
}

// Exposed Method to Fetch Personalised Feed
func (handlers *FeedHandlers) FetchPersonalisedFeed(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	apiKey := headers[utils.HeadersApiKey]
	memberRole := headers[utils.HeadersMemberRole]

	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	isCm := utils.IsCMRole(memberRole)

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	loggedInUser := LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// Fetch user personalised feed from cache
	cacheKey := fmt.Sprintf(cache.UserPersonalisedFeedKey, communityId, userId)
	userPersonalisedFeed, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	postIds := []string{}

	if exists {
		err := json.Unmarshal([]byte(userPersonalisedFeed), &postIds)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

	} else { // If user personalised feed not found, fetch community default feed

		cacheKey = fmt.Sprintf(cache.CommunityDefaultFeedKey, communityId)
		defaultFeed, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if exists {
			err := json.Unmarshal([]byte(defaultFeed), &postIds)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

		} else {

			// Fetch community bot id
			botId := externalHelpers.GetCommunityBotId(apiKey, "")

			// compute and save community default feed
			computeAndSaveCommunityDefaultFeed(handlers, communityId, botId)

			cacheKey = fmt.Sprintf(cache.CommunityDefaultFeedKey, communityId)
			defaultFeed, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			if exists {
				err := json.Unmarshal([]byte(defaultFeed), &postIds)
				if err != nil {
					utils.GeneralAPIInternalError(c, err.Error())
					return
				}
			} else {
				utils.GeneralAPIValidationError(c, "Personalised & Default feed is not yet computed. Please try again later.")
				return
			}
		}
	}

	// fetch pagination query params
	page, pageSize, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// slice postIds based on pagination
	startIndex := (page - 1) * pageSize
	endIndex := int(math.Min(float64(len(postIds)), float64(page*pageSize)))

	postIds = postIds[startIndex:endIndex]
	postObjectIds := helpers.ConvertIdsToObjectIds(postIds)

	// Fetch posts data from post service
	postFilter := gin.H{
		"_id": gin.H{
			"$in": postObjectIds,
		},
		"is_deleted":   false,
		"community_id": communityId,
	}

	postsData, err := handlers.postHelper.FindPostHelper(postFilter, nil)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// sort the posts based on postIds order
	postsMap := make(map[primitive.ObjectID]entities.Post, len(postsData))
	for _, postData := range postsData {
		postsMap[postData.ID] = postData
	}

	sortedPosts := []entities.Post{}
	for _, postId := range postObjectIds {
		if post, ok := postsMap[postId]; ok {
			sortedPosts = append(sortedPosts, post)
		}
	}

	// parse posts for multiple post response
	parsedPosts := parseMultiplePostResponse(handlers, sortedPosts, userId, isCm, versionCode, platformCode, apiRevampV1Check, memberRole)

	// parse posts for final response (topics, widgets, comments, etc)
	finalParsedResponse := parsePostsAndGenerateFinalResponse(handlers, &loggedInUser, parsedPosts)

	// return final response
	utils.GenerateSuccessResponse(c, finalParsedResponse)
}

// Exposed Method to Reorder Personalised Feed
func (handlers *FeedHandlers) ReorderPersonalisedFeed(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	MemberRole := headers[utils.HeadersMemberRole]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	isCm := utils.IsCMRole(MemberRole)

	// Reorder user personalised feed and update in cache
	reorderUserPersonalisedFeed(handlers, communityId, userId, isCm) // TODO: Can move this to background service

	utils.GenerateSuccessResponse(c, nil)
}

// method to reorder user personalised feed and update in cache
func reorderUserPersonalisedFeed(handlers *FeedHandlers, communityId int, userId string, isCm bool) {

	// fetch community metric post scores
	postScoreMap, err := fetchCommunityMetricPostScores(handlers, communityId, userId)
	if err != nil {
		logging.Error("Error in fetching post metrics for community: ", communityId, " err: ", err)
		return
	}

	if len(postScoreMap) == 0 {
		logging.Error("No community metric post scores found for community: ", communityId)
		return
	}

	// fetch user specific metric scores map
	userSpecificMetricScores, err := fetchUserSpecificMetricScores(handlers, userId, communityId)
	if err != nil {
		logging.Error("Error in fetching user specific metric scores: ", err)
		return
	}

	// Add user specific metric scores to posts score map
	for postId, score := range userSpecificMetricScores {
		postScoreMap[postId] += score
	}

	// Sort the post score map in descending order and get top 1000 posts
	sortedPostIds := utils.SortFloatMapByValues(postScoreMap, true)
	if len(sortedPostIds) > 1000 {
		sortedPostIds = sortedPostIds[:1000]
	}

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in fetching block user list from cache", err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// Filter for all posts of community
	allPostsOfCommunityFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"community_id": communityId,
				"is_deleted":   false,
				"$and": []gin.H{
					{
						"$or": []gin.H{
							{
								"user_id": gin.H{"$in": excludedUserIds},
							},
						},
					},
				},
			},
		},
		gin.H{
			"$project": gin.H{
				"_id":        1,
				"created_at": 1,
			},
		},
	}

	// Add hidden post filter for non-CM users
	if !isCm {
		hiddenPostFilter := gin.H{
			"is_hidden": true,
			"user_id":   gin.H{"$ne": userId},
		}

		allPostsOfCommunityFilter[0]["$match"].(gin.H)["$and"].([]gin.H)[0]["$or"] = append(allPostsOfCommunityFilter[0]["$match"].(gin.H)["$and"].([]gin.H)[0]["$or"].([]gin.H), hiddenPostFilter)
	}

	toExcludePostData, err := handlers.postHelper.AggregatePostHelper(allPostsOfCommunityFilter)
	if err != nil {
		logging.Error("Error in fetching all blocked user posts of community: ", communityId, " err: ", err)
		return
	}

	toExcludePostIds := []string{}
	for _, postData := range toExcludePostData {
		toExcludePostIds = append(toExcludePostIds, postData["_id"].(primitive.ObjectID).Hex())
	}

	sortedPostIds = utils.GetDifferenceBetweenStringArray(sortedPostIds, toExcludePostIds)

	// Save user personalised feed (postIds) in cache
	cacheKey := fmt.Sprintf(cache.UserPersonalisedFeedKey, communityId, userId)
	defaultFeedBytesValue, _ := json.Marshal(sortedPostIds)

	setStatus := handlers.cacheHelper.Set(cacheKey, defaultFeedBytesValue, cache.UserPersonalisedFeedCacheTTLInHours*time.Minute)
	if setStatus.Err() != nil {
		logging.Error("Error in saving user personalised feed in cache", setStatus.Err())
	}
}

// Internal method to compute and save community default feed in cache
func computeAndSaveCommunityDefaultFeed(handlers *FeedHandlers, communityId int, userId string) error {
	// Fetch post scores map for the community
	postScoreMap, err := fetchCommunityMetricPostScores(handlers, communityId, userId)
	if err != nil {
		logging.Error("Error in fetching post metrics for community: ", communityId, " err: ", err)
		return err
	}

	if len(postScoreMap) == 0 {
		return nil
	}

	// Sort the post score map in descending order and get top 1000 posts
	sortedPostIds := utils.SortFloatMapByValues(postScoreMap, true)
	if len(sortedPostIds) > 1000 {
		sortedPostIds = sortedPostIds[:1000]
	}

	// Save the default community feed in cache
	cacheKey := fmt.Sprintf(cache.CommunityDefaultFeedKey, communityId)
	defaultFeedBytesValue, _ := json.Marshal(sortedPostIds)

	setStatus := handlers.cacheHelper.Set(cacheKey, defaultFeedBytesValue, cache.DefaultCommunityFeedCacheTTLInMins*time.Minute)
	if setStatus.Err() != nil {
		logging.Error("Error in saving community default feed in cache", setStatus.Err())
		return err
	}

	return nil
}

// Exposed Method to compute community default feed | Should be run every 30 mins
func (handlers *FeedHandlers) ComputeCommunityDefaultFeed(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Call compute and save community default feed
	computeAndSaveCommunityDefaultFeed(handlers, communityId, userId)

	utils.GenerateSuccessResponse(c, nil)
}

// Exposed Method to compute community default feed in async every 30 mins
func AsyncComputeCommunityDefaultFeed(handlers *FeedHandlers) error {
	communityIdsList := externalHelpers.GetCommunityIdsForCommunitySettingsEnabled(handlers.cacheHelper, externalHelpers.PersonalisedFeedSettingType)

	for _, communityId := range communityIdsList {
		userId := externalHelpers.GetCommunityBotId("", communityId)

		communityIdInt, err := strconv.Atoi(communityId)
		if err != nil {
			logging.Error(fmt.Sprintf("Error in converting the community id: %s to int", communityId))
			continue
		}

		// Call compute and save community default feed
		computeAndSaveCommunityDefaultFeed(handlers, communityIdInt, userId)
	}

	return nil
}

func fetchCommunityMetricPostScores(handlers *FeedHandlers, communityId int, userId string,
) (map[string]float64, error) {

	postScoreMap := map[string]float64{}

	// Fetch all the recent posts of the community
	recentPostsMapData := map[string]float64{}
	cacheKey := fmt.Sprintf(cache.PostsRecencyMetricsKey, communityId)
	recentPostsMap, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return postScoreMap, fmt.Errorf("error in fetching community recency metrics from cache: %v", err)
	}

	if exists {
		err := json.Unmarshal([]byte(recentPostsMap), &recentPostsMapData)
		if err != nil {
			return postScoreMap, fmt.Errorf("error in unmarshalling recency metrics from cache: %v", err)
		}
	} else { // Compute recency metric if not found in cache
		logging.Info("Recency metrics not found in cache. Computing recency metrics for community: ", communityId)
		recentPostsMapData = RecencyMetricComputation(handlers, userId, communityId)
	}

	for postId := range recentPostsMapData {
		postScoreMap[postId] += recentPostsMapData[postId]
	}

	// Fetch all the top liked posts of the community
	topLikedPostsMapData := map[string]float64{}
	cacheKey = fmt.Sprintf(cache.PostsLikesMetricsKey, communityId)
	topLikedPostsMap, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return postScoreMap, fmt.Errorf("error in fetching community likes metrics from cache: %v", err)
	}

	if exists {
		err := json.Unmarshal([]byte(topLikedPostsMap), &topLikedPostsMapData)
		if err != nil {
			return postScoreMap, fmt.Errorf("error in unmarshalling likes metrics from cache: %v", err)
		}
	} else {
		logging.Info("Likes metrics not found in cache. Computing likes metrics for community: ", communityId)
		topLikedPostsMapData = PostLikesMetricComputation(handlers, userId, communityId)
	}

	for postId := range topLikedPostsMapData {
		postScoreMap[postId] += topLikedPostsMapData[postId]
	}

	// Fetch all the top commented posts of the community
	topCommentedPostsMapData := map[string]float64{}
	cacheKey = fmt.Sprintf(cache.PostsCommentsMetricsKey, communityId)
	topCommentedPostsMap, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return postScoreMap, fmt.Errorf("error in fetching community comments metrics from cache: %v", err)
	}
	if exists {
		err := json.Unmarshal([]byte(topCommentedPostsMap), &topCommentedPostsMapData)
		if err != nil {
			return postScoreMap, fmt.Errorf("error in unmarshalling comments metrics from cache: %v", err)
		}
	} else {
		logging.Info("Comments metrics not found in cache. Computing comments metrics for community: ", communityId)
		topCommentedPostsMapData = PostCommentsMetricComputation(handlers, userId, communityId)
	}

	for postId := range topCommentedPostsMapData {
		postScoreMap[postId] += topCommentedPostsMapData[postId]
	}

	return postScoreMap, nil
}

func fetchUserSpecificMetricScores(handlers *FeedHandlers, userId string, communityId int,
) (map[string]float64, error) {

	userSpecificMetricScores := map[string]float64{}

	// fetch user groups metric score map for user
	cacheKey := fmt.Sprintf(cache.UserGroupsMetricsKey, communityId, userId)
	userGroupsMetricMapCacheValue, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return nil, fmt.Errorf("error in fetching user groups metric score from cache: %v", err)
	}

	var userGroupsMetricMap map[string]float64
	if exists {
		err := json.Unmarshal([]byte(userGroupsMetricMapCacheValue), &userGroupsMetricMap)
		if err != nil {
			return nil, fmt.Errorf("error in unmarshalling user groups metric score from cache: %v", err)
		}

		for postId, score := range userGroupsMetricMap {
			userSpecificMetricScores[postId] += score
		}
	}

	// fetch user topics metric score map for user
	cacheKey = fmt.Sprintf(cache.UserTopicsMetricsKey, communityId, userId)
	userTopicsMetricMapCacheValue, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return nil, fmt.Errorf("error in fetching user topics metric score from cache: %v", err)
	}

	var userTopicsMetricMap map[string]float64
	if exists {
		err := json.Unmarshal([]byte(userTopicsMetricMapCacheValue), &userTopicsMetricMap)
		if err != nil {
			return nil, fmt.Errorf("error in unmarshalling user topics metric score from cache: %v", err)
		}

		for postId, score := range userTopicsMetricMap {
			userSpecificMetricScores[postId] += score
		}
	}

	// fetch user dampened posts score map
	userDampenedPostsScoreMap, err := UserDampenedMetricsComputation(handlers, userId, communityId)
	if err != nil {
		return nil, fmt.Errorf("error in fetching user dampened posts: %v", err)
	}

	// reduce the score of dampened posts
	for postId, score := range userDampenedPostsScoreMap {
		userSpecificMetricScores[postId] -= score
	}

	// fetch user connection metric score map for user
	cacheKey = fmt.Sprintf(cache.UserConnectionMetricsKey, communityId, userId)
	userConnectionMetricMapCacheValue, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return nil, fmt.Errorf("error in fetching user connection metric score from cache: %v", err)
	}

	var userConnectionMetricMap map[string]float64
	if exists {
		err := json.Unmarshal([]byte(userConnectionMetricMapCacheValue), &userConnectionMetricMap)
		if err != nil {
			return nil, fmt.Errorf("error in unmarshalling user groups metric score from cache: %v", err)
		}

		for postId, score := range userConnectionMetricMap {
			userSpecificMetricScores[postId] += score
		}
	}

	return userSpecificMetricScores, nil
}

// Internal method to compute and save recency metric for newly created post
func updateRecencyMetricForNewlycreatedPost(postHelper interfaces.PostHelper, cacheHelper cache.Helper, postId string) error {

	// Fetch post using postId
	postfilter := gin.H{
		"_id":        postId,
		"is_deleted": false,
	}

	postData, err := postHelper.FindPostHelper(postfilter, nil)
	if err != nil {
		return fmt.Errorf("error in fetching post: %v", err)
	}

	if len(postData) == 0 {
		return fmt.Errorf("post not found with postId: %s", postId)
	}

	// Get posts recency metrics data from cache
	cacheKey := fmt.Sprintf(cache.PostsRecencyMetricsKey, postData[0].CommunityId)
	postsMetricMapCacheValue, exists, err := cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		return fmt.Errorf("error in fetching posts recency metrics from cache: %v", err)
	}

	if !exists {
		logging.Info("Posts recency metrics not present in cache for community: ", postData[0].CommunityId)
		return nil
	}

	// Update the post recency metric score with the new post score in cache
	var postRecencyMetricsMap map[string]float64
	json.Unmarshal([]byte(postsMetricMapCacheValue), &postRecencyMetricsMap)

	botId := externalHelpers.GetCommunityBotId("", fmt.Sprint(postData[0].CommunityId))
	if botId == "" {
		return fmt.Errorf("error in fetching botId")
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(cacheHelper, botId, postData[0].CommunityId)
	if err != nil {
		return fmt.Errorf("error in fetching personalised feed weights: %v", err)
	}

	// Compute recency metric score for post
	postCreatedAtInSeconds := float64(postData[0].CreatedAt.Unix())
	currentTimeInSeconds := float64(time.Now().Unix())

	metricScore := 0.0

	if personalisedFeedWeights.RecencyMetrics.MaxThreshold-(currentTimeInSeconds-postCreatedAtInSeconds) > 0 {
		metricScore = computeRecencyMetricScore(
			postCreatedAtInSeconds, personalisedFeedWeights.RecencyMetrics.MaxThreshold,
			personalisedFeedWeights.RecencyMetrics.Weight, currentTimeInSeconds)
	}

	postRecencyMetricsMap[postId] = metricScore

	// Set the updated post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(postRecencyMetricsMap)
	setStatus := cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.CommunityMetricCacheTTLInMins*time.Minute)
	if setStatus.Err() != nil {
		logging.Error("Error in saving recency metric score in cache", setStatus.Err())
	}

	return nil
}
