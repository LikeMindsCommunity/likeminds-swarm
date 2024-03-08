package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createFilterQueryToGetTopicIdsBasedOnTopicsFilter(topicIds []string) []map[string]interface{} {
	postIdsFilterData := []map[string]interface{}{}

	// Add group query
	postIdsFilterData = append(postIdsFilterData, gin.H{
		"$group": bson.M{
			"_id": "$post_id",
			"topics": bson.M{
				"$addToSet": "$topic_id",
			},
		},
	})

	// Build and add match query
	var topicsWithOrList [][]primitive.ObjectID
	var filterQueryList []gin.H

	for _, topicId := range topicIds {
		var topicsObjectIdWithAndList []primitive.ObjectID

		topicsWithAndList := strings.Split(string(topicId), constants.TopicsSplitterKeyWithAndValue)
		if len(topicsWithAndList) > 0 {
			topicsObjectIdWithAndList = helpers.ConvertIdsToObjectIds(topicsWithAndList)
			topicsWithOrList = append(topicsWithOrList, topicsObjectIdWithAndList)
		}

		filterQuery := gin.H{
			"topics": gin.H{
				"$all": topicsObjectIdWithAndList,
			},
		}

		filterQueryList = append(filterQueryList, filterQuery)
	}

	postIdsFilterData = append(postIdsFilterData, bson.M{"$match": bson.M{"$or": filterQueryList}})

	// Add group query
	postIdsFilterData = append(postIdsFilterData, gin.H{
		"$group": bson.M{
			"_id": "",
			"post_ids": bson.M{
				"$addToSet": "$_id",
			},
		},
	})

	// Add unset query
	postIdsFilterData = append(postIdsFilterData, gin.H{
		"$unset": "_id",
	})

	return postIdsFilterData
}

func getPostIdsBasedOnTopicsFilter(handlers *FeedHandlers, topicIds []string) ([]primitive.ObjectID, error) {
	var postIds []primitive.ObjectID

	postIdsFilterData := createFilterQueryToGetTopicIdsBasedOnTopicsFilter(topicIds)

	// Fetch post ids using helper method
	postIdsResult, err := handlers.postTopicsHelper.AggregatePostTopicsHelper(postIdsFilterData)
	if err != nil {
		return nil, err
	}

	if len(postIdsResult) > 0 {
		postIds = postIdsResult[0].PostIDs
	}

	return postIds, nil
}
