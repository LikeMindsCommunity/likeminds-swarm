package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommentRepository interface {
	Create(like *entities.Comment) (interface{}, error)
	Find(filter map[string]interface{}, filterOptions *options.FindOptions) ([]entities.Comment, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

type CommentHelper interface {
	CreateCommentHelper(text string, postId primitive.ObjectID, level int, userId string) (interface{}, error)
	FindCommentHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Comment, error)
	UpdateCommentByIdHelper(comment_id primitive.ObjectID, update map[string]interface{}) error
	CountCommentHelper(filter map[string]interface{}) (int64, error)
}
