package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Comment Repository
type CommentRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOptions *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	UpdateMany(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) (*mongo.Cursor, error)
}

// Interface for Comment Helper
type CommentHelper interface {
	CreateCommentHelper(text string, postId primitive.ObjectID, communityId int, level int, userId string,
		tempId *string, createdAt int, attachments []requests.AttachmentRequest) (interface{}, error)
	FindCommentHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Comment, error)
	EditCommentHelper(commentId primitive.ObjectID, text string, attachments []requests.AttachmentRequest, markIsEdited bool) error
	UpdateCommentByIdHelper(commentId primitive.ObjectID, update map[string]interface{}) error
	UpdateManyCommentsHelper(filter map[string]interface{}, update map[string]interface{}) error
	CountCommentHelper(filter map[string]interface{}) (int64, error)
	AggregateTopCommentsHelper(query []map[string]interface{}) ([]responses.TopCommentsAggregationQueryResponse, error)
}
