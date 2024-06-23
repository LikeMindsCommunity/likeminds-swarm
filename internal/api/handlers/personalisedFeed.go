package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Method to Add a New Poll Option
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
	PostLikesMetricComputation(handlers, userId, communityId)

	// Compute post comments metric and save it in cache
	PostCommentsMetricComputation(handlers, userId, communityId)

	// Compute user groups metric and save it in cache
	UserGroupsMetricComputation(handlers, userId, communityId, apiKey)

	// Compute user topics metric and save it in cache
	UserTopicsMetricComputation(handlers, userId, communityId)

	utils.GenerateSuccessResponse(c, gin.H{})
}

// Recompute & save post ranking on the basis of recency metrics
func RecencyMetricComputation(handlers *FeedHandlers, userId string, communityId int) {
	postsMetricMap := map[string]float64{}

	cacheKey := fmt.Sprintf(cache.PostsRececnyMetricsKey, communityId)

	// Get data from cache
	postsMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey)
	if postsMetricMapCacheValue.Val() != "" && postsMetricMapCacheValue.Val() != "null" {
		return
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)

	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return
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

	currentTimeInSeconds := float64(time.Now().Unix())

	for _, postData := range allPostsData {
		var metricScore float64

		postCreatedAtInSeconds := float64(postData["created_at"].(primitive.DateTime).Time().Unix())
		if personalisedFeedWeights.RecencyMetrics.MaxThreshold-(float64(time.Now().Unix())-postCreatedAtInSeconds) > 0 {
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
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.PostsRecenctCacheTTLInMins*time.Minute)

	if setStatus.Err() != nil {
		logging.Error("Error in saving recency metric score in cache", setStatus.Err())
	}
}

// Compute recency metric score
func computeRecencyMetricScore(postCreatedAt float64, recencyMetricMaxThreshold float64, recencyMetricWeight float64, currentTime float64) float64 {
	return ((recencyMetricMaxThreshold - (float64(time.Now().Unix()) - postCreatedAt)) / recencyMetricMaxThreshold) * recencyMetricWeight
}

// Recompute & save post ranking on the basis of likes metrics
func PostLikesMetricComputation(handlers *FeedHandlers, userId string, communityId int) {
	postsLikesMetricMap := map[string]float64{}

	cacheKey := fmt.Sprintf(cache.PostsLikesMetricsKey, communityId)

	// Get data from cache
	postsLikesMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey)
	if postsLikesMetricMapCacheValue.Val() != "" && postsLikesMetricMapCacheValue.Val() != "null" {
		return
	}

	// Start of computation of post likes metric
	var postRecencyMetricsMap map[string]float64
	postIdsArray := []string{}
	postsMetricMapCacheKey := fmt.Sprintf(cache.PostsRececnyMetricsKey, communityId)

	// Get data from cache
	postsMetricMapCacheValue := handlers.cacheHelper.Get(postsMetricMapCacheKey)
	if postsMetricMapCacheValue.Val() == "" || postsMetricMapCacheValue.Val() == "null" {
		return
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
		return
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
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.PostsRecenctCacheTTLInMins*time.Minute)

	if setStatus.Err() != nil {
		logging.Error("Error in saving post likes metric score in cache", setStatus.Err())
	}
}

// Compute post likes metric score
func computePostLikesMetricScore(postLikesCount float64, likesMetricMaxThreshold float64, likesMetricWeight float64) float64 {
	return utils.GetMinimumFromArray(postLikesCount, likesMetricMaxThreshold) * likesMetricWeight
}

// Recompute & save post ranking on the basis of comments metrics
func PostCommentsMetricComputation(handlers *FeedHandlers, userId string, communityId int) {
	postsCommentsMetricMap := map[string]float64{}

	cacheKey := fmt.Sprintf(cache.PostsCommentsMetricsKey, communityId)

	// Get data from cache
	postsCommentsMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey)
	if postsCommentsMetricMapCacheValue.Val() != "" && postsCommentsMetricMapCacheValue.Val() != "null" {
		return
	}

	// Get personalised weights
	personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	if err != nil {
		logging.Error("Error in computation of recency metric: ", err)
		return
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
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.PostsRecenctCacheTTLInMins*time.Minute)

	if setStatus.Err() != nil {
		logging.Error("Error in saving post comments metric score in cache", setStatus.Err())
	}
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

	if allUserFollowedChannelPostsData == nil || len(allUserFollowedChannelPostsData) == 0 {
		logging.Error("No user followed channels post data found in db")
		return
	}

	for _, postData := range allUserFollowedChannelPostsData {
		var metricScore float64

		metricScore = computeUserGroupsMetricScore(
			personalisedFeedWeights.UserGroupsMetrics.MaxThreshold,
			personalisedFeedWeights.UserGroupsMetrics.Weight)

		userGroupsMetricMap[postData["_id"].(primitive.ObjectID).Hex()] = metricScore
	}

	// Set post metric score in cache
	postsMetricMapBytesValue, _ := json.Marshal(userGroupsMetricMap)

	cacheKey := fmt.Sprintf(cache.UserGroupsMetricsKey, userId, communityId)
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

	allUserFollowedChannelPostsData, _ := handlers.userTopicsHelper.AggregateUserTopicsHelper(allUserTopicsPostsOfCommunityFilter)

	if allUserFollowedChannelPostsData == nil || len(allUserFollowedChannelPostsData) == 0 {
		logging.Error("No user followed channels post data found in db")
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

	cacheKey := fmt.Sprintf(cache.UserTopicsMetricsKey, userId, communityId)
	setStatus := handlers.cacheHelper.Set(cacheKey, postsMetricMapBytesValue, cache.UserMetricCacheTTLInHours*time.Hour)

	if setStatus.Err() != nil {
		logging.Error("Error in saving user groups metric score in cache", setStatus.Err())
	}
}

// Compute user topics metric score
func computeUserTopicsMetricScore(topicsCount float64, userTopicsMetricMaxThreshold float64, userTopicsMetricWeight float64) float64 {
	return utils.GetMinimumFromArray(topicsCount, userTopicsMetricMaxThreshold) * userTopicsMetricWeight
}
