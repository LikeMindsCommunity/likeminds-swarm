package interfaces

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserEntityTimestampRepository interface {
	Create(document interface{}) (interface{}, error)
	CreateMany(documents []interface{}) ([]interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	UpdateMany(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	DeleteMany(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) (*mongo.Cursor, error)
}

type UserEntityTimestampHelper interface {
	CreateUserEntityTimestampHelper(userId string, communityId int, entityType string, entityIds []primitive.ObjectID, epochTimeStamp int) ([]primitive.ObjectID, error)
	FindUserEntityTimestampHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.UserTopic, error)
	UpdateManyUserEntityTimestampHelper(filter map[string]interface{}, update map[string]interface{}) error
	CountUserEntityTimestampHelper(filter map[string]interface{}) (int64, error)
	DeleteUserEntityTimestampHelper(filter map[string]interface{}) error
	AggregateUserEntityTimestampHelper(query []map[string]interface{}) ([]gin.H, error)
}
