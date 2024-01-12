package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Pending Post
func (helper *pendingPostHelper) CreatePendingPostHelper(text string, heading string, communityId int, userId string,
	attachments []requests.Attachment, chatroomId int, tempId *string, topicIds []primitive.ObjectID, originalAuthorUUID string,
	visibility string, isRepost bool, createdAt int, postType string) (interface{}, error) {

	// parse attachments
	postAttachments := parseAttachments(attachments)

	if tempId != nil && *tempId == "" {
		tempId = nil
	}

	pendingPost := entities.NewPendingPost(text, heading, communityId, userId, postAttachments, chatroomId, tempId,
		topicIds, originalAuthorUUID, visibility, isRepost, createdAt, postType)
	id, err := helper.pendingPostRepository.Create(&pendingPost)

	return id, err
}

// Exposed Helper Method to Edit Post
func (helper pendingPostHelper) EditPendingPostHelper(id primitive.ObjectID, text string, heading string, attachments []requests.Attachment,
	topicIds []primitive.ObjectID, visibility string, markIsEdited bool, postType string) error {

	// parse attachments
	postAttachments := parseAttachments(attachments)

	updateBody := gin.H{
		"$set": gin.H{
			"post_data._id":         id,
			"post_data.text":        text,
			"post_data.heading":     heading,
			"post_data.attachments": postAttachments,
			"post_data.topic_ids":   topicIds,
			"post_data.visibility":  visibility,
			"post_type":             postType,
			"updated_at":            time.Now(),
		},
	}

	if markIsEdited {
		updateBody["$set"].(gin.H)["is_edited"] = markIsEdited
	}

	err := helper.pendingPostRepository.Update(gin.H{"_id": id}, updateBody)

	return err

}

// Exposed Helper Method to Find Pending Post
func (helper pendingPostHelper) FindPendingPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) (
	[]entities.PendingPost, error) {

	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	cursor, err := helper.pendingPostRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.PendingPost
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Pending Post
func (helper *pendingPostHelper) UpdatePendingPostByIdHelper(id primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	err := helper.pendingPostRepository.Update(gin.H{"_id": id}, update)

	return err
}

// Exposed Helper Method to Fetch Pending Posts Count
func (helper pendingPostHelper) CountPendingPostHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.pendingPostRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to perform Aggregration on Posts
func (helper pendingPostHelper) AggregatePendingPostHelper(query []map[string]interface{}) ([]gin.H, error) {
	for _, value := range query {
		if matchGroup, ok := value["$match"]; ok {
			err := convertHexIdsToObjectIds(matchGroup.(gin.H), []string{"_id", "entity_id"})
			if err != nil {
				return nil, err
			}
		}
	}

	results, err := helper.pendingPostRepository.Aggregate(query)

	return results, err
}

// Structure for Pending Post Helper
type pendingPostHelper struct {
	pendingPostRepository interfaces.PendingPostRepository
}

// Exposed Method to Create New Post Helper
func NewPendingPostHelper(pendingPostRepository interfaces.PendingPostRepository) interfaces.PendingPostHelper {
	return &pendingPostHelper{
		pendingPostRepository: pendingPostRepository,
	}
}
