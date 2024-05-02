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

// internal method to parse attachments
func parseAttachments(attachments []requests.AttachmentRequest) []entities.Attachment {

	var parsedAttachments []entities.Attachment

	// parse attachments
	for _, element := range attachments {

		metaData := element.AttachmentMeta
		ogTags := metaData.OgTags
		metaOgTags := entities.NewOgTags(ogTags.Title, ogTags.Image, ogTags.Description, ogTags.Url)

		entityId := primitive.NilObjectID
		if metaData.EntityID != "" {
			entityId, _ = primitive.ObjectIDFromHex(metaData.EntityID)
		}

		attachmentMeta := entities.NewAttachmentMeta(metaData.Name, metaData.Url, metaData.Format, metaData.Size, metaData.Duration,
			metaData.PageCount, metaData.ThumbnailUrl, metaOgTags, entityId, metaData.CoverImageUrl, metaData.Title, metaData.Body,
			metaData.ExpiryTime, metaData.PollType, metaData.MultipleSelectState, metaData.MultipleSelectNumber, metaData.IsAnonymous,
			metaData.AllowAddOption, metaData.NsfwScore, metaData.Height, metaData.Width)

		attachment := entities.NewAttachment(element.AttachmentType, attachmentMeta)
		parsedAttachments = append(parsedAttachments, attachment)
	}

	return parsedAttachments
}

// Exposed Helper Method to Create Post
func (helper *postHelper) CreatePostHelper(text string, heading string, communityId int, userId string, attachments []requests.AttachmentRequest,
	chatroomId int, tempId *string, topicIds []primitive.ObjectID, originalAuthorUUID string, visibility string, isRepost bool, createdAt int) (interface{}, error) {

	// parse attachments
	postAttachments := parseAttachments(attachments)

	if tempId != nil && *tempId == "" {
		tempId = nil
	}

	post := entities.NewPost(text, heading, communityId, userId, postAttachments, chatroomId, tempId, topicIds, originalAuthorUUID, visibility, isRepost, createdAt)
	postId, err := helper.postRepository.Create(&post)

	return postId, err
}

// Exposed Helper Method to Edit Post
func (helper *postHelper) EditPostHelper(postId primitive.ObjectID, text string, heading string, attachments []requests.AttachmentRequest,
	topicIds []primitive.ObjectID, visibility string, markIsEdited bool) error {

	// parse attachments
	postAttachments := parseAttachments(attachments)

	updateBody := gin.H{
		"$set": gin.H{
			"text":        text,
			"heading":     heading,
			"attachments": postAttachments,
			"topic_ids":   topicIds,
			"visibility":  visibility,
			"updated_at":  time.Now(),
		},
	}

	if markIsEdited {
		updateBody["$set"].(gin.H)["is_edited"] = markIsEdited
	}

	err := helper.postRepository.Update(gin.H{"_id": postId}, updateBody)

	return err

}

// Exposed Helper Method to Find Post
func (helper *postHelper) FindPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Post, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	cursor, err := helper.postRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Post
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Post
func (helper *postHelper) UpdatePostByIdHelper(postId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	err := helper.postRepository.Update(gin.H{"_id": postId}, update)

	return err
}

// Exposed Helper Method to Update Multiple Posts
func (helper *postHelper) UpdateManyPostsHelper(filter map[string]interface{}, update map[string]interface{}, shouldUpdateTimestamp bool) error {
	if shouldUpdateTimestamp {
		if _, ok := update["$set"]; ok {
			update["$set"].(gin.H)["updated_at"] = time.Now()
		} else {
			update["$set"] = gin.H{
				"updated_at": time.Now(),
			}
		}
	}

	return helper.postRepository.UpdateMany(filter, update)
}

// Exposed Helper Method to Fetch Post Count
func (helper *postHelper) CountPostHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.postRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to perform Aggregration on Posts
func (helper *postHelper) AggregatePostHelper(query []map[string]interface{}) ([]gin.H, error) {
	for _, value := range query {
		if matchGroup, ok := value["$match"]; ok {
			err := convertHexIdsToObjectIds(matchGroup.(gin.H), []string{"_id", "entity_id"})
			if err != nil {
				return nil, err
			}
		}
	}

	results, err := helper.postRepository.Aggregate(query)

	return results, err
}

// Structure for Post Helper
type postHelper struct {
	postRepository interfaces.PostRepository
}

// Exposed Method to Create New Post Helper
func NewPostHelper(postRepository interfaces.PostRepository) interfaces.PostHelper {
	return &postHelper{
		postRepository: postRepository,
	}
}
