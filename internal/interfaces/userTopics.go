package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserTopicsRepository interface {
	Create(document interface{}) (interface{}, error)
	CreateMany(documents []interface{}) ([]interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	UpdateMany(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	DeleteMany(filter map[string]interface{}) (int64, error)
}

type UserTopicsHelper interface {
	CreateUsersTopicsHelper(usersTopicIds map[string][]primitive.ObjectID, communityId int) ([]primitive.ObjectID, error)
	FindUserTopicsHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.UserTopic, error)
	UpdateManyUserTopicsHelper(filter map[string]interface{}, update map[string]interface{}) error
	CountUserTopicsHelper(filter map[string]interface{}) (int64, error)
	DeleteUserTopicsHelper(filter map[string]interface{}) error
}
