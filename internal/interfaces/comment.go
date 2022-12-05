package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentRepository interface {
	Create(like *entities.Comment) (interface{}, error)
	Find(filter map[string]interface{}) ([]entities.Comment, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
}

type CommentHelper interface {
	CreateCommentHelper(text string, postId primitive.ObjectID, level int, userId string) (interface{}, error)
	FindCommentHelper(filter map[string]interface{}) ([]entities.Comment, error)
	UpdateCommentHelper(filter map[string]interface{}, update map[string]interface{}) error
}
