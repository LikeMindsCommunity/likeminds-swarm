package handlers

import (
	"fmt"
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
	orQueryList, notQueryList := []gin.H{}, []gin.H{}

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
					objectId, err := primitive.ObjectIDFromHex(topicsWithOnlyList[1])
					if err != nil {
						fmt.Printf("Error converting to ObjectID: %v\n", err)
						continue
					}

					filterQuery = gin.H{
						"original_topics": gin.H{
							"$eq": objectId,
						},
					}
				}
			case constants.TopicsSplitterKeyWithNotValue:
				topicsWithNotList := strings.Split(topicId, constants.TopicsSplitterKeyWithNotValue)
				if len(topicsWithNotList) > 1 {
					objectId, err := primitive.ObjectIDFromHex(topicsWithNotList[1])
					if err != nil {
						fmt.Printf("Error converting to ObjectID: %v\n", err)
						continue
					}

					notQuery := gin.H{
						"topics": gin.H{
							"$not": bson.M{
								"$eq": objectId,
							},
						},
					}

					notQueryList = append(notQueryList, notQuery)
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
			orQueryList = append(orQueryList, filterQuery)
		}

	}

	finalFilterQuery := bson.M{}

	if len(orQueryList) > 0 {
		finalFilterQuery["$or"] = orQueryList
	}

	if len(notQueryList) > 0 {
		finalFilterQuery["$and"] = notQueryList
	}

	postIdsFilterData = append(postIdsFilterData, bson.M{"$match": finalFilterQuery})

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
