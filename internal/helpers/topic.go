package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
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
