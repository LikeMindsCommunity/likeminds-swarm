package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LikeRepository interface {
	Create(like *entities.Like) (interface{}, error)
	Find(filter map[string]interface{}) ([]entities.Like, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

type LikeHelper interface {
	CreateLikeHelper(entity_type string, entity_id primitive.ObjectID, liked_by string) (interface{}, error)
	FindLikeHelper(filter map[string]interface{}) ([]entities.Like, error)
	UpdateLikeHelper(filter map[string]interface{}, update map[string]interface{}) error
	CountLikeHelper(filter map[string]interface{}) (int64, error)
}
