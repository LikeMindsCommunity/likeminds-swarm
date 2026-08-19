package interfaces

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/responses"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Post Topics Repository
type PostTopicsRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	CreateorUpdateMany(filterListWithUpdate [][]gin.H) error
	Count(filter map[string]interface{}) (int64, error)
	DeleteMany(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) (*mongo.Cursor, error)
}

// Interface for Post Topics Helper
type PostTopicsHelper interface {
	CreatePostTopicsHelper(postId primitive.ObjectID, topicId primitive.ObjectID, IsOriginalTopic bool, community_id int) (interface{}, error)
	FindPostTopicsHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.PostTopic, error)
	CreateOrUpdateManyPostTopicsHelper(postId primitive.ObjectID, originalTopicIds []primitive.ObjectID, parentTopicIds []primitive.ObjectID, communityId int) error
	UpdatePostTopicsByIdHelper(topicId primitive.ObjectID, update map[string]interface{}) error
	CountPostTopicsHelper(filter map[string]interface{}) (int64, error)
	DeletePostTopicsHelper(filter gin.H) error
	AggregatePostTopicsHelper(query []map[string]interface{}) ([]responses.PostIdsBasedonTopics, error)
}
