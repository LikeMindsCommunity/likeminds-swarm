package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Pending Post Repository
type PendingPostRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) ([]gin.H, error)
}

// Interface for Pending Post Helper
type PendingPostHelper interface {
	CreatePendingPostHelper(text string, heading string, communityId int, userId string, attachments []requests.AttachmentRequest,
		chatroomId int, tempId *string, topicIds []primitive.ObjectID, originalAuthorUUID string, visibility string,
		isRepost bool, createdAt int, postType string) (interface{}, error)
	EditPendingPostHelper(id primitive.ObjectID, text string, heading string, attachments []requests.AttachmentRequest,
		topicIds []primitive.ObjectID, visibility string, markIsEdited bool, postType string) error
	FindPendingPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.PendingPost, error)
	UpdatePendingPostByIdHelper(id primitive.ObjectID, update map[string]interface{}) error
	CountPendingPostHelper(filter map[string]interface{}) (int64, error)
	AggregatePendingPostHelper(query []map[string]interface{}) ([]gin.H, error)
}
