package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
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
	UpdateMany(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	DeleteMany(filter map[string]interface{}) (int64, error)
}

// Interface for Topic Helper
type TopicHelper interface {
	CreateTopicHelper(name string, is_enabled bool, priority float32, isSearchable bool, parentId primitive.ObjectID, parentName string, allParentIds []primitive.ObjectID, level int, widgetId primitive.ObjectID, totalChildCount int, access string, communityId int) (interface{}, error)
	CreateManyTopicsHelper(topicsRequest []requests.CreateTopicRequest, communityId int) ([]primitive.ObjectID, error)
	FindTopicHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Topic, error)
	UpdateTopicByIdHelper(topicId primitive.ObjectID, update map[string]interface{}, updateTimestamp bool) error
	UpdateManyTopicsHelper(filter map[string]interface{}, update map[string]interface{}, shouldUpdateTimestamp bool) error
	CountTopicHelper(filter map[string]interface{}) (int64, error)
	DeleteTopicsHelper(topicIds []primitive.ObjectID) error
}
