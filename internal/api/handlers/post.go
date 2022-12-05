package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func (handlers *postHandlers) CreatePost(c *gin.Context) {
	// validation of request body
	var createPostRequest requests.CreatePostRequest
	if err := c.ShouldBindJSON(&createPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch headers
	headers := utils.GetHeaders(c)

	// validation of attachment objects
	for _, element := range createPostRequest.Attachments {

		switch element.FileType {
		case constants.ImageWidget:
			if element.FileUrl == "" {
				utils.GeneralAPIValidationError(c, "send file_url in attachment")
				return
			}

		case constants.VideoWidget:
			if element.FileUrl == "" {
				utils.GeneralAPIValidationError(c, "send file_url in attachment")
				return
			}

		case constants.DocumentWidget:
			if element.FileUrl == "" {
				utils.GeneralAPIValidationError(c, "send file_url in attachment")
				return
			}

			if element.FileFormat == "" {
				utils.GeneralAPIValidationError(c, "send file_format in attachment")
				return
			}

			if element.FileSize == "" {
				utils.GeneralAPIValidationError(c, "send file_size in attachment")
				return
			}

		default:
			utils.GeneralAPIValidationError(c, "send valid file_type in attachment")
			return
		}
	}

	// create post using the helper method
	err := handlers.postHelper.CreatePostHelper(createPostRequest.Text, headers[utils.HeadersApiKey],
		headers[utils.HeadersMemberId], createPostRequest.Attachments)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *postHandlers) FetchPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	post_data, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"post":    post_data,
	})
}

func (handlers *postHandlers) DeletePost(c *gin.Context) {
	// validation of request body
	var deletePostRequest requests.DeletePostRequest
	if err := c.ShouldBindJSON(&deletePostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post using helper method
	post_data, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deletePostRequest.UserIsCm && headers[utils.HeadersMemberId] != post_data.UserId {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// update data
	update_data := gin.H{
		"is_deleted":    true,
		"delete_reason": deletePostRequest.DeleteReason,
		"deleted_by":    headers[utils.HeadersMemberId],
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(post_data.ID, update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// TODO-create activity of post deletion

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *postHandlers) PinPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post using helper method
	post_data, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(post_data.ID, gin.H{"is_pinned": true})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

type postHandlers struct {
	postHelper interfaces.PostHelper
}

func NewPostHandlers(postHelper interfaces.PostHelper) *postHandlers {
	return &postHandlers{postHelper: postHelper}
}
