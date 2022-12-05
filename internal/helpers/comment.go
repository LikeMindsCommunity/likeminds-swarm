package helpers

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (helper *commentHelper) CreateCommentHelper(text string, postId primitive.ObjectID, level int, userId string) (interface{}, error) {
	comment := entities.NewComment(text, postId, level, userId)
	comment_id, err := helper.commentRepository.Create(&comment)

	return comment_id, err
}

func (helper *commentHelper) FindCommentHelper(filter map[string]interface{}) ([]entities.Comment, error) {
	results, err := helper.commentRepository.Find(filter)

	return results, err
}

func (helper *commentHelper) UpdateCommentHelper(filter map[string]interface{}, update map[string]interface{}) error {
	err := helper.commentRepository.Update(filter, update)

	return err
}

type commentHelper struct {
	commentRepository interfaces.CommentRepository
}

func NewCommentHelper(commentRepository interfaces.CommentRepository) interfaces.CommentHelper {
	return &commentHelper{
		commentRepository: commentRepository,
	}
}
