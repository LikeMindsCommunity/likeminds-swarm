package handlers

import (
	"regexp"
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
			"original_topics": bson.M{
				"$addToSet": bson.M{
					"$cond": bson.A{
						"$is_original_topic",
						"$topic_id",
						nil,
					},
				},
			},
		},
	})

	// Build and add match query
	var filterQueryList []gin.H

	// Define regex pattern for special delimiters
	delimiterPattern := regexp.MustCompile(`#\$(.*?)\$#`)

	for _, topicId := range topicIds {

		filterQuery := gin.H{}

		// Check if topic contains any special delimiters
		match := delimiterPattern.FindString(topicId)
		if match != "" {

			switch match {
			case constants.TopicsSplitterKeyWithAndValue:
				topicsWithAndList := strings.Split(topicId, constants.TopicsSplitterKeyWithAndValue)
				if len(topicsWithAndList) > 0 {
					filterQuery = gin.H{
						"topics": gin.H{
							"$all": helpers.ConvertIdsToObjectIds(topicsWithAndList),
						},
					}
				}
			case constants.TopicsSplitterKeyWithOnlyValue:
				topicsWithOnlyList := strings.Split(topicId, constants.TopicsSplitterKeyWithOnlyValue)
				if len(topicsWithOnlyList) > 1 {
					objectId, _ := primitive.ObjectIDFromHex(topicsWithOnlyList[1])

					filterQuery = gin.H{
						"original_topics": gin.H{
							"$eq": objectId,
						},
					}
				}
			}
		} else {
			filterQuery = gin.H{
				"topics": gin.H{
					"$all": helpers.ConvertIdsToObjectIds([]string{topicId}),
				},
			}
		}

		if len(filterQuery) > 0 {
			filterQueryList = append(filterQueryList, filterQuery)
		}

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
