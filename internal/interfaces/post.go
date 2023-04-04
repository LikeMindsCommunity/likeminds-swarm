package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Post Repository
type PostRepository interface {
	Create(post *entities.Post) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Post, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) ([]gin.H, error)
}

// Interface for Post Helper
type PostHelper interface {
	CreatePostHelper(text string, heading string, community_id int, user_id string, attachments []requests.Attachment, chatroom_id int) (interface{}, error)
	EditPostHelper(post_id primitive.ObjectID, text string, heading string, attachments []requests.Attachment) error
	FindPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Post, error)
	UpdatePostByIdHelper(post_id primitive.ObjectID, update map[string]interface{}) error
	CountPostHelper(filter map[string]interface{}) (int64, error)
	AggregatePostHelper(query []map[string]interface{}) ([]gin.H, error)
}
