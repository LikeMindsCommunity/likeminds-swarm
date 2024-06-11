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

	// Filter for all posts of community
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
