package interfaces

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Connection Feed Repository
type ConnectionFeedRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) (interface{}, error)
}

// Interface for Connection Feed Helper
type ConnectionFeedHelper interface {
	CreateConnectionFeedHelper(postId primitive.ObjectID, userId string, communityId int) (interface{}, error)
	FindConnectionFeedHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.ConnectionFeed, error)
}
