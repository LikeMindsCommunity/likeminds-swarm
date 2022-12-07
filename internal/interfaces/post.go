package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostRepository interface {
	Create(post *entities.Post) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Post, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

type PostHelper interface {
	CreatePostHelper(text string, api_key string, user_id string, attachments []requests.Attachment) (interface{}, error)
	FindPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Post, error)
	UpdatePostByIdHelper(post_id primitive.ObjectID, update map[string]interface{}) error
	CountPostHelper(filter map[string]interface{}) (int64, error)
}
