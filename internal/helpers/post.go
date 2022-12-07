package helpers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (helper *postHelper) CreatePostHelper(text string, api_key string, user_id string, attachments []requests.Attachment) error {
	var post_attachments []interface{}

	for _, element := range attachments {

		switch element.FileType {
		case constants.ImageWidget:
			image_widget := entities.NewImageWidget(element.FileType, element.FileUrl)
			post_attachments = append(post_attachments, image_widget)

		case constants.VideoWidget:
			video_widget := entities.NewVideoWidget(element.FileType, element.FileUrl)
			post_attachments = append(post_attachments, video_widget)

		case constants.DocumentWidget:
			document_widget := entities.NewDocumentWidget(element.FileType, element.FileUrl, element.FileFormat, element.FileSize)
			post_attachments = append(post_attachments, document_widget)
		}

	}

	post := entities.NewPost(text, api_key, user_id, post_attachments)
	err := helper.postRepository.Create(&post)

	return err
}

func (helper *postHelper) FindPostHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Post, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertMultipleHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	results, err := helper.postRepository.Find(filter, &fOpts)

	return results, err
}

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

func (helper *postHelper) CountPostHelper(filter map[string]interface{}) (int64, error) {
	err := convertMultipleHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.postRepository.Count(filter)

	return count, err
}

type postHelper struct {
	postRepository interfaces.PostRepository
}

func NewPostHelper(postRepository interfaces.PostRepository) interfaces.PostHelper {
	return &postHelper{
		postRepository: postRepository,
	}
}
