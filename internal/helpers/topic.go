package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Topic Instance
func (helper *topicHelper) CreateTopicHelper(name string, is_enabled bool, community_id int) (interface{}, error) {
	// Create a new Topic Document
	topic := entities.NewTopic(name, is_enabled, community_id)

	// Insert the document in the collection
	topicId, err := helper.topicRepository.Create(topic)

	return topicId, err
}

// Exposed Helper Method to Create Topic Instances
func (helper *topicHelper) CreateManyTopicsHelper(names []string, is_enabled bool, community_id int) ([]interface{}, error) {
	// Create new topic documents
	var topics []interface{}
	for _, name := range names {
		topics = append(topics, entities.NewTopic(name, is_enabled, community_id))
	}

	// Insert the documents in the collection
	topicIds, err := helper.topicRepository.CreateMany(topics)

	return topicIds, err
}

// Exposed Helper Method to Find Topics
func (helper *topicHelper) FindTopicHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Topic, error) {
	fOpts := mergeFilterOptions(filterOptions)

	// Parse the object Ids
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.topicRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Topic
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Topics
func (helper *topicHelper) UpdateTopicByIdHelper(topic_id primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	// Create set filter
	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	// Update the document in the collection
	err := helper.topicRepository.Update(gin.H{"_id": topic_id}, update)

	return err
}

func (helper *topicHelper) DeleteTopicsHelper(topic_ids []primitive.ObjectID) error {
	// create a filter to delete topic instances with topic_ids
	filter := gin.H{
		"_id": gin.H{
			"$in": topic_ids,
		},
	}

	// Delete the documents from the collection
	deletedCount, err := helper.topicRepository.DeleteMany(filter)

	// log count if all documents were not deleted
	if deletedCount != int64(len(topic_ids)) {
		logging.Error(fmt.Sprintf(`
			Deleted %d out of %d topics using DeleteMany.
			`, deletedCount, len(topic_ids)),
		)
	}

	return err
}

// Exposed Helper Method to Fetch Topics Count
func (helper *topicHelper) CountTopicHelper(filter map[string]interface{}) (int64, error) {
	// Parse the object IDs
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return 0, err
	}

	// Get count of documents from the collection
	count, err := helper.topicRepository.Count(filter)

	return count, err
}

// Structure for Topic Helper
type topicHelper struct {
	topicRepository interfaces.TopicRepository
}

// Exposed Method to Create New Topic Helper
func NewTopicHelper(topicRepository interfaces.TopicRepository) interfaces.TopicHelper {
	return &topicHelper{
		topicRepository: topicRepository,
	}
}
