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
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

// Interface for Topic Helper
type TopicHelper interface {
	CreateTopicHelper(name string, is_enabled bool, community_id int) (interface{}, error)
	FindTopicHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Topic, error)
	UpdateTopicByIdHelper(topic_id primitive.ObjectID, update map[string]interface{}) error
	CountTopicHelper(filter map[string]interface{}) (int64, error)
}
