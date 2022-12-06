package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LikeRepository interface {
	Create(like *entities.Like) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Like, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	CountById(filter map[string]interface{}) (int64, error)
	Aggregate(query []interface{}) (interface{}, error)
}

type LikeHelper interface {
	CreateLikeHelper(entity_type string, entity_id primitive.ObjectID, liked_by string) (interface{}, error)
	FindLikeHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Like, error)
	UpdateLikeByIdHelper(like_id primitive.ObjectID, update map[string]interface{}) error
	CountLikeHelper(filter map[string]interface{}) (int64, error)
	AggregateLikeHelper(query []interface{}) (interface{}, error)
}
