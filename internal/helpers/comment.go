package helpers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (helper *commentHelper) CreateCommentHelper(text string, postId primitive.ObjectID, level int, userId string) (interface{}, error) {
	comment := entities.NewComment(text, postId, level, userId)
	comment_id, err := helper.commentRepository.Create(&comment)

	return comment_id, err
}

func (helper *commentHelper) FindCommentHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Comment, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertMultipleHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return nil, err
	}
	results, err := helper.commentRepository.Find(filter, &fOpts)

	return results, err
}

func (helper *commentHelper) UpdateCommentByIdHelper(comment_id primitive.ObjectID, update map[string]interface{}) error {
	set_data := gin.H{}

	if _, ok := update["$set"]; ok {
		set_data = update["$set"].(gin.H)
	}
	set_data["updated_at"] = time.Now()
	update["$set"] = set_data

	err := helper.commentRepository.Update(gin.H{"_id": comment_id}, update)

	return err
}

func (helper *commentHelper) CountCommentHelper(filter map[string]interface{}) (int64, error) {
	err := convertMultipleHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.commentRepository.Count(filter)

	return count, err
}

type commentHelper struct {
	commentRepository interfaces.CommentRepository
}

func NewCommentHelper(commentRepository interfaces.CommentRepository) interfaces.CommentHelper {
	return &commentHelper{
		commentRepository: commentRepository,
	}
}
