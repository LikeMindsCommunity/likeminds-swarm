package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create a Comment
func (helper *commentHelper) CreateCommentHelper(text string, postId primitive.ObjectID, communityId int, level int, userId string,
	tempId *string, createdAt int) (interface{}, error) {
	if tempId != nil && *tempId == "" {
		tempId = nil
	}

	comment := entities.NewComment(text, postId, communityId, level, userId, tempId, createdAt)
	commentId, err := helper.commentRepository.Create(&comment)

	return commentId, err
}

// Exposed Helper Method to Find a Comment
func (helper *commentHelper) FindCommentHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Comment, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.commentRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Comment
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update a Comment
func (helper *commentHelper) UpdateCommentByIdHelper(commentId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	err := helper.commentRepository.Update(gin.H{"_id": commentId}, update)

	return err
}

// Exposed Helper Method to Count Comments
func (helper *commentHelper) CountCommentHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.commentRepository.Count(filter)

	return count, err
}

// Structure for Comment Helper
type commentHelper struct {
	commentRepository interfaces.CommentRepository
}

// Exposed Method to Create a New Comment Helper
func NewCommentHelper(commentRepository interfaces.CommentRepository) interfaces.CommentHelper {
	return &commentHelper{
		commentRepository: commentRepository,
	}
}
