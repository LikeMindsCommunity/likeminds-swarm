package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
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
