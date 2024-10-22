package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Post Topics Instance
func (helper *postTopicsHelper) CreatePostTopicsHelper(postId primitive.ObjectID, topicId primitive.ObjectID,
	IsOriginalTopic bool, community_id int,
) (interface{}, error) {

	// Create a new Topic Document
	postTopic := entities.NewPostTopic(postId, topicId, IsOriginalTopic, community_id)

	// Insert the document in the collection
	postTopicId, err := helper.postTopicsRepository.Create(postTopic)

	return postTopicId, err
}

// Exposed Helper Method to Create or Update Many Post Topics Instances
func (helper *postTopicsHelper) CreateOrUpdateManyPostTopicsHelper(
	postId primitive.ObjectID, originalTopicIds []primitive.ObjectID, parentTopicIds []primitive.ObjectID, communityId int,
) error {

	if len(originalTopicIds) > 0 {
		filterWithMapList := createPostTopicsFilter(postId, originalTopicIds, true, communityId)

		err := helper.postTopicsRepository.CreateorUpdateMany(filterWithMapList)
		if err != nil {
			logging.Error(fmt.Sprintf(`Error in CreateorUpdateMany for original topics. Error: %v`, err))
			return err
		}
	}

	if len(parentTopicIds) > 0 {
		filterWithMapList := createPostTopicsFilter(postId, parentTopicIds, false, communityId)

		err := helper.postTopicsRepository.CreateorUpdateMany(filterWithMapList)
		if err != nil {
			logging.Error(fmt.Sprintf(`Error in CreateorUpdateMany for parent topics. Error: %v`, err))
			return err
		}
	}

	return nil
}

func createPostTopicsFilter(postId primitive.ObjectID, topicIds []primitive.ObjectID, IsOriginalTopics bool, communityId int,
) [][]gin.H {

	var filterWithMapList [][]gin.H

	for _, topicId := range topicIds {
		var filterWithUpdate []gin.H

		filter := gin.H{
			"post_id":      postId,
			"topic_id":     topicId,
			"community_id": communityId,
			"created_at": gin.H{
				"$exists": true,
			},
			"updated_at": gin.H{
				"$exists": true,
			},
		}

		update := gin.H{
			"$set": gin.H{
				"created_at":        time.Now(),
				"updated_at":        time.Now(),
				"is_original_topic": IsOriginalTopics,
			},
		}

		filterWithUpdate = append(filterWithUpdate, filter, update)
		filterWithMapList = append(filterWithMapList, filterWithUpdate)
	}

	return filterWithMapList
}

// Exposed Helper Method to Find Post Topics
func (helper *postTopicsHelper) FindPostTopicsHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.PostTopic, error) {
	fOpts := mergeFilterOptions(filterOptions)

	// Parse the object Ids
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.postTopicsRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.PostTopic
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Many Post Topics
func (helper *postTopicsHelper) UpdatePostTopicsByIdHelper(postTopicId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	// Create set filter
	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	// Update the document in the collection
	err := helper.postTopicsRepository.Update(gin.H{"_id": postTopicId}, update)

	return err
}

func (helper *postTopicsHelper) DeletePostTopicsHelper(filter gin.H) error {
	// Delete the documents from the collection
	deletedCount, err := helper.postTopicsRepository.DeleteMany(filter)

	// log count if all documents were not deleted
	logging.Error(fmt.Sprintf(`Deleted %d post topics using DeleteMany.`, deletedCount))

	return err
}

// Exposed Helper Method to Fetch Post Topics Count
func (helper *postTopicsHelper) CountPostTopicsHelper(filter map[string]interface{}) (int64, error) {
	// Parse the object IDs
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return 0, err
	}

	// Get count of documents from the collection
	count, err := helper.postTopicsRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to perform Aggregration on Posts
func (helper *postTopicsHelper) AggregatePostTopicsHelper(query []map[string]interface{}) ([]responses.PostIdsBasedonTopics, error) {
	results, err := helper.postTopicsRepository.Aggregate(query)
	if err != nil {
		logging.Error(fmt.Sprintf(`Error in aggregation of postTopics. Error: %v`, err))
		return nil, err
	}

	var postIdsList []responses.PostIdsBasedonTopics

	if err = results.All(context.TODO(), &postIdsList); err != nil {
		return postIdsList, fmt.Errorf("Error in conversion!")
	}

	return postIdsList, nil
}

// Structure for Post Topics Helper
type postTopicsHelper struct {
	postTopicsRepository interfaces.PostTopicsRepository
}

// Exposed Method to Create New Post Topics Helper
func NewPostTopicsHelper(postTopicRepository interfaces.PostTopicsRepository) interfaces.PostTopicsHelper {
	return &postTopicsHelper{
		postTopicsRepository: postTopicRepository,
	}
}
