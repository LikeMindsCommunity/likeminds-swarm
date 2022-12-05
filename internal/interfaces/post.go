package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostRepository interface {
	Create(post *entities.Post) error
	Find(filter map[string]interface{}) ([]entities.Post, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
}

type PostHelper interface {
	CreatePostHelper(text string, api_key string, user_id string, attachments []requests.Attachment) error
	FindPostByIdHelper(post_id string, api_key string) (*entities.Post, error)
	FindPostHelper(filter map[string]interface{}) ([]entities.Post, error)
	UpdatePostByIdHelper(post_id primitive.ObjectID, update map[string]interface{}) error
	UpdatePostHelper(filter map[string]interface{}, update map[string]interface{}) error
}
