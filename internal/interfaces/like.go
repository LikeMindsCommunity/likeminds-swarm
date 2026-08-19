package interfaces

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Like Repository
type LikeRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) (interface{}, error)
}

// Interface for Like Helper
type LikeHelper interface {
	CreateLikeHelper(entityType string, entityId primitive.ObjectID, likedBy string, CreatedAt int) (interface{}, error)
	FindLikeHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Like, error)
	UpdateLikeByIdHelper(likeId primitive.ObjectID, update map[string]interface{}) error
	CountLikeHelper(filter map[string]interface{}) (int64, error)
	AggregateLikeHelper(query []map[string]interface{}) (interface{}, error)
}
