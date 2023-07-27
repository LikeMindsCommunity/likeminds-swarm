package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Post Repository
type PostRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) ([]gin.H, error)
}

// Interface for Post Helper
type PostHelper interface {
	CreatePostHelper(text string, heading string, communityId int, userId string, attachments []requests.Attachment,
		chatroomId int, tempId *string, topicIds []primitive.ObjectID) (interface{}, error)
	EditPostHelper(post_id primitive.ObjectID, text string, heading string, attachments []requests.Attachment,
		topicIds []primitive.ObjectID) error
	FindPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Post, error)
	UpdatePostByIdHelper(post_id primitive.ObjectID, update map[string]interface{}) error
	CountPostHelper(filter map[string]interface{}) (int64, error)
	AggregatePostHelper(query []map[string]interface{}) ([]gin.H, error)
}
