package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Topic Repository
type TopicRepository interface {
	Create(document interface{}) (interface{}, error)
	CreateMany(documents []interface{}) ([]interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	DeleteMany(filter map[string]interface{}) error
}

// Interface for Topic Helper
type TopicHelper interface {
	CreateTopicHelper(name string, isEnabled bool, communityId int) (interface{}, error)
	CreateManyTopicsHelper(names []string, isEnabled bool, communityId int) ([]interface{}, error)
	FindTopicHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Topic, error)
	UpdateTopicByIdHelper(topicId primitive.ObjectID, update map[string]interface{}) error
	CountTopicHelper(filter map[string]interface{}) (int64, error)
	DeleteTopicsHelper(topicIds []primitive.ObjectID) error
}
