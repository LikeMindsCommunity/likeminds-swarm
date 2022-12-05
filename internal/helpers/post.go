package helpers

import (
	"fmt"

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

func (helper *postHelper) FindPostByIdHelper(post_id string, api_key string) (*entities.Post, error) {
	// post filter data
	post_filter_data := gin.H{
		"_id":        post_id,
		"is_deleted": false,
		"api_key":    api_key,
	}

	// fetch post using helper method
	post_results, err := helper.FindPostHelper(post_filter_data)
	if err != nil {
		return nil, err
	}

	// validation of post_id
	if len(post_results) == 0 {
		return nil, fmt.Errorf("invalid post_id sent")
	}

	return &post_results[0], nil
}

func (helper *postHelper) FindPostHelper(filter map[string]interface{}) ([]entities.Post, error) {
	results, err := helper.postRepository.Find(filter)

	return results, err
}

func (helper *postHelper) UpdatePostByIdHelper(post_id primitive.ObjectID, update map[string]interface{}) error {
	err := helper.UpdatePostHelper(gin.H{"_id": post_id}, update)
	if err != nil {
		return err
	}

	return nil
}

func (helper *postHelper) UpdatePostHelper(filter map[string]interface{}, update map[string]interface{}) error {
	err := helper.postRepository.Update(filter, update)

	return err
}

type postHelper struct {
	postRepository interfaces.PostRepository
}

func NewPostHelper(postRepository interfaces.PostRepository) interfaces.PostHelper {
	return &postHelper{
		postRepository: postRepository,
	}
}
