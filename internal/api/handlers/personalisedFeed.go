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
	postsLikesMetricMapCacheValue := handlers.cacheHelper.Get(cacheKey) // Can use GetWithKeyExists to check if key exists
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

	// // Start of computation of post likes metric
	// var postRecencyMetricsMap map[string]float64
	// postIdsArray := []string{}
	// postsMetricMapCacheKey := fmt.Sprintf(cache.PostsRececnyMetricsKey, communityId)

	// // Get data from cache
	// postsMetricMapCacheValue := handlers.cacheHelper.Get(postsMetricMapCacheKey)
	// if postsMetricMapCacheValue.Val() == "" || postsMetricMapCacheValue.Val() == "null" {
	// 	return
	// }

	// err := json.Unmarshal([]byte(postsMetricMapCacheValue.Val()), &postRecencyMetricsMap)

	// if err != nil {
	// 	logging.Error("Error in unmarshalling recency metric score from cache", err)
	// }

	// for postId := range postRecencyMetricsMap {
	// 	postIdsArray = append(postIdsArray, postId)
	// }

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

// Exposed Method to Fetch Personalised Feed
func (handlers *FeedHandlers) FetchPersonalisedFeed(c *gin.Context) {

	// fetch headers and url params
	// headers := utils.GetHeaders(c)
	// userId := headers[utils.HeadersMemberId]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Get personalised weights
	// personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	// if err != nil {
	// 	utils.GeneralAPIInternalError(c, fmt.Sprint("Error in fetching personalised feed weights: ", err.Error()))
	// 	return
	// }

	// Get personalised feed
	// personalisedFeed, err := externalHelpers.GetPersonalisedFeed(handlers.cacheHelper, userId, communityId, personalisedFeedWeights)
	// if err != nil {
	// utils.GeneralAPIInternalError(c, fmt.Sprint("Error in fetching personalised feed: ", err.Error()))
	// return
	// }

	// response := gin.H{}

	utils.GenerateSuccessResponse(c, gin.H{})
}

// Exposed Method to Reorder Personalised Feed
func (handlers *FeedHandlers) ReorderPersonalisedFeed(c *gin.Context) {

	// fetch headers and url params
	// headers := utils.GetHeaders(c)
	// userId := headers[utils.HeadersMemberId]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// Get personalised weights
	// personalisedFeedWeights, err := externalHelpers.GetPersonalisedFeedWeightsAgainstCommunity(handlers.cacheHelper, userId, communityId)
	// if err != nil {
	// 	utils.GeneralAPIInternalError(c, fmt.Sprint("Error in fetching personalised feed weights: ", err.Error()))
	// 	return
	// }

	// Get personalised feed
	// personalisedFeed, err := externalHelpers.GetPersonalisedFeed(handlers.cacheHelper, userId, communityId, personalisedFeedWeights)
	// if err != nil {
	// 	utils.GeneralAPIInternalError(c, fmt.Sprint("Error in fetching personalised feed: ", err.Error()))
	// 	return
	// }

	// Reorder personalised feed
	// reorderedPersonalisedFeed, err := externalHelpers.ReorderPersonalisedFeed(personalisedFeed)
	// if err != nil {
	// 	utils.GeneralAPIInternalError(c, fmt.Sprint("Error in reordering personalised feed: ", err.Error()))
	// 	return
	// }

	// response := gin.H{}

	utils.GenerateSuccessResponse(c, gin.H{})
}

// Exposed Method to compute community default feed | Should be run every 30 mins
func (handlers *FeedHandlers) ComputeCommunityDefaultFeed(communityId int) {

	postScoreMap := map[string]float64{}

	// Fetch all the recent posts of the community
	cacheKey := fmt.Sprintf(cache.PostsRececnyMetricsKey, communityId)
	recentPostsMap, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		logging.Error("Error in fetching community recency metrics from cache", err)
		return
	}

	if exists {
		var recentPostsMapData map[string]float64
		err := json.Unmarshal([]byte(recentPostsMap), &recentPostsMapData)
		if err != nil {
			logging.Error("Error in unmarshalling recency metrics from cache", err)
			return
		}

		for postId := range recentPostsMapData {
			postScoreMap[postId] += recentPostsMapData[postId]
		}
	}

	// Fetch all the top liked posts of the community
	cacheKey = fmt.Sprintf(cache.PostsLikesMetricsKey, communityId)
	topLikedPostsMap, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		logging.Error("Error in fetching community likes metrics from cache", err)
		return
	}

	if exists {
		var topLikedPostsMapData map[string]float64
		err := json.Unmarshal([]byte(topLikedPostsMap), &topLikedPostsMapData)
		if err != nil {
			logging.Error("Error in unmarshalling likes metrics from cache", err)
			return
		}

		for postId := range topLikedPostsMapData {
			postScoreMap[postId] += topLikedPostsMapData[postId]
		}
	}

	// Fetch all the top commented posts of the community
	cacheKey = fmt.Sprintf(cache.PostsCommentsMetricsKey, communityId)
	topCommentedPostsMap, exists, err := handlers.cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		logging.Error("Error in fetching community comments metrics from cache", err)
		return
	}

	if exists {
		var topCommentedPostsMapData map[string]float64
		err := json.Unmarshal([]byte(topCommentedPostsMap), &topCommentedPostsMapData)
		if err != nil {
			logging.Error("Error in unmarshalling comments metrics from cache", err)
			return
		}

		for postId := range topCommentedPostsMapData {
			postScoreMap[postId] += topCommentedPostsMapData[postId]
		}
	}

	if len(postScoreMap) == 0 {
		logging.Error("No post metrics found for community: ", communityId)
		return
	}

	// Sort the post score map in descending order and get top 1000 posts
	sortedPostIds := utils.SortMapByValues(postScoreMap, true)
	if len(sortedPostIds) > 1000 {
		sortedPostIds = sortedPostIds[:1000]
	}

	// Save the default community feed in cache
	cacheKey = fmt.Sprintf(cache.CommunityDefaultFeedKey, communityId)
	defaultFeedBytesValue, _ := json.Marshal(sortedPostIds)

	setStatus := handlers.cacheHelper.Set(cacheKey, defaultFeedBytesValue, cache.DefaultCommunityFeedCacheTTLInMins*time.Minute)
	if setStatus.Err() != nil {
		logging.Error("Error in saving community default feed in cache", setStatus.Err())
	}

}
