package helpers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Post
func (helper *postHelper) CreatePostHelper(text string, heading string, community_id int, user_id string, attachments []requests.Attachment, chatroom_id int) (interface{}, error) {
	var post_attachments []entities.Attachment

	for _, element := range attachments {

		meta_data := element.AttachmentMeta
		og_tags := meta_data.OgTags
		meta_og_tags := entities.NewOgTags(og_tags.Title, og_tags.Image, og_tags.Description, og_tags.Url)
		attachment_meta := entities.NewAttachmentMeta(meta_data.Name, meta_data.Url, meta_data.Format, meta_data.Size, meta_data.Duration, meta_data.PageCount, meta_og_tags)
		attachment := entities.NewAttachment(element.AttachmentType, attachment_meta)
		post_attachments = append(post_attachments, attachment)

	}

	post := entities.NewPost(text, heading, community_id, user_id, post_attachments, chatroom_id)
	post_id, err := helper.postRepository.Create(&post)

	return post_id, err
}

// Exposed Helper Method to Find Post
func (helper *postHelper) FindPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Post, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	results, err := helper.postRepository.Find(filter, &fOpts)

	return results, err
}

// Exposed Helper Method to Update Post
func (helper *postHelper) UpdatePostByIdHelper(post_id primitive.ObjectID, update map[string]interface{}) error {
	set_data := gin.H{}

	if _, ok := update["$set"]; ok {
		set_data = update["$set"].(gin.H)
	}
	set_data["updated_at"] = time.Now()
	update["$set"] = set_data

	err := helper.postRepository.Update(gin.H{"_id": post_id}, update)

	return err
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
