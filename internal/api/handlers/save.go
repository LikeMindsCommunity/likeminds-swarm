package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func fetchPostIdsFromSave(saved_posts []entities.Save) []primitive.ObjectID {
	post_id_list := []primitive.ObjectID{}
	for _, save := range saved_posts {
		if save.EntityType == constants.PostEntityType {
			post_id_list = append(post_id_list, save.EntityId)
		}
	}

	return post_id_list
}

func fetchUserSavedPostByPostId(helper interfaces.SaveHelper, post_id string, saved_by string) ([]entities.Save, error) {
	// save filter data
	save_filter_data := gin.H{
		"entity_id":   post_id,
		"entity_type": constants.PostEntityType,
		"saved_by":    saved_by,
		"is_deleted":  false,
	}

	// fetch save using helper method
	save_results, err := helper.FindSaveHelper(save_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	return save_results, nil
}

func fetchUserSavedStatusByPostId(helper interfaces.SaveHelper, post_id string, saved_by string) bool {
	save_results, err := fetchUserSavedPostByPostId(helper, post_id, saved_by)
	if err != nil {
		return false
	}

	if len(save_results) == 0 {
		return false
	}

	return true
}

func (handlers *saveHandlers) SavePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch save using helper method
	save_results, err := fetchUserSavedPostByPostId(handlers.saveHelper, post_id, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(save_results) == 0 {
		// create save using the helper method
		_, err := handlers.saveHelper.CreateSaveHelper(constants.PostEntityType, post_data.ID,
			headers[utils.HeadersMemberId])
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	} else {
		save_data := save_results[0]

		// save update data
		save_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !save_data.IsDeleted,
			},
		}

		// update save using the helper method
		err = handlers.saveHelper.UpdateSaveByIdHelper(save_data.ID, save_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *saveHandlers) FetchUserSavedPosts(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	user_id := c.Param("user_id")

	if user_id != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// save filter data
	save_filter_data := gin.H{
		"entity_type": constants.PostEntityType,
		"saved_by":    headers[utils.HeadersMemberId],
		"is_deleted":  false,
	}

	// fetch save count using helper method
	save_count, err := handlers.saveHelper.CountSaveHelper(save_filter_data)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// filter options
	save_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch save using helper method
	save_results, err := handlers.saveHelper.FindSaveHelper(save_filter_data, save_filter_options)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// posts filter data
	post_filter_data := gin.H{
		"_id": gin.H{
			"$in": fetchPostIdsFromSave(save_results),
		},
		"is_deleted": false,
	}

	// fetch posts using helper method
	post_results, err := handlers.postHelper.FindPostHelper(post_filter_data, save_filter_options)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	saved_post_response := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper,
		post_results, user_id)

	// return final response
	c.JSON(http.StatusOK, parseFetchMultiplePostResponse(handlers.postHelper, saved_post_response, save_count))
}

type saveHandlers struct {
	saveHelper    interfaces.SaveHelper
	likeHelper    interfaces.LikeHelper
	commentHelper interfaces.CommentHelper
	postHelper    interfaces.PostHelper
}

func NewSaveHandlers(saveHelper interfaces.SaveHelper, likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	postHelper interfaces.PostHelper) *saveHandlers {
	return &saveHandlers{
		saveHelper:    saveHelper,
		likeHelper:    likeHelper,
		commentHelper: commentHelper,
		postHelper:    postHelper,
	}
}
