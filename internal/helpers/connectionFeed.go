package helpers

import (
	"context"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create a Connection Feed
func (helper *connectionFeedHelper) CreateConnectionFeedHelper(postId primitive.ObjectID, userId string, communityId int) (interface{}, error) {
	connectionFeedItem := entities.NewConnectionFeedItem(postId, userId, communityId)
	connectionFeedItemId, err := helper.connectionFeedRepository.Create(&connectionFeedItem)

	return connectionFeedItemId, err
}

// Exposed Helper Method to Find Connection Feed
func (helper *connectionFeedHelper) FindConnectionFeedHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.ConnectionFeed, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.connectionFeedRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.ConnectionFeed
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Structure for Connection Feed Helper
type connectionFeedHelper struct {
	connectionFeedRepository interfaces.ConnectionFeedRepository
}

// Exposed Method to create New Connection Feed Helper
func NewConnectionFeedHelper(connectionFeedRepository interfaces.ConnectionFeedRepository) interfaces.ConnectionFeedHelper {
	return &connectionFeedHelper{
		connectionFeedRepository: connectionFeedRepository,
	}
}
