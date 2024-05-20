package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create a Comment
func (helper *commentHelper) CreateCommentHelper(text string, postId primitive.ObjectID, communityId int, level int, userId string,
	tempId *string, createdAt int, attachments []requests.AttachmentRequest) (interface{}, error) {

	if tempId != nil && *tempId == "" {
		tempId = nil
	}

	// parse attachments
	parsedAttachments := parseAttachments(attachments)

	comment := entities.NewComment(text, postId, communityId, level, userId, tempId, createdAt, parsedAttachments)
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

// Exposed Helper Method to Edit a Comment
func (helper *commentHelper) EditCommentHelper(commentId primitive.ObjectID, text string, attachments []requests.AttachmentRequest,
	markIsEdited bool) error {

	// parse attachments
	parsedAttachments := parseAttachments(attachments)

	updateBody := gin.H{
		"$set": gin.H{
			"text":        text,
			"attachments": parsedAttachments,
			"updated_at":  time.Now(),
		},
	}

	if markIsEdited {
		updateBody["$set"].(gin.H)["is_edited"] = true
	}

	err := helper.commentRepository.Update(gin.H{"_id": commentId}, updateBody)

	return err

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

// Exposed Helper Method to Update Many Comments
func (helper *commentHelper) UpdateManyCommentsHelper(filter map[string]interface{}, update map[string]interface{}) error {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return err
	}

	err = helper.commentRepository.UpdateMany(filter, update)

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

// Exposed Helper Method to perform Aggregation on Comments
func (helper *commentHelper) AggregateTopCommentsHelper(query []map[string]interface{}) ([]responses.TopCommentsAggregationQueryResponse, error) {

	results, err := helper.commentRepository.Aggregate(query)

	var commentResultsList []responses.TopCommentsAggregationQueryResponse
	if err = results.All(context.TODO(), &commentResultsList); err != nil {
		return nil, fmt.Errorf("Error in conversion!")
	}

	return commentResultsList, err
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
