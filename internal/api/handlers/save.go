package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func (handlers *saveHandlers) SavePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post using helper method
	post_data, err := handlers.postHelper.FindPostByIdHelper(post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// save filter data
	save_filter_data := gin.H{
		"entity_id":   post_data.ID,
		"entity_type": constants.PostEntityType,
		"saved_by":    headers[utils.HeadersMemberId],
	}

	// fetch save using helper method
	save_results, err := handlers.saveHelper.FindSaveHelper(save_filter_data)
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
		err = handlers.saveHelper.UpdateSaveHelper(save_filter_data, save_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// TODO- create activity of save

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

type saveHandlers struct {
	saveHelper interfaces.SaveHelper
	postHelper interfaces.PostHelper
}

func NewSaveHandlers(saveHelper interfaces.SaveHelper, postHelper interfaces.PostHelper) *saveHandlers {
	return &saveHandlers{
		saveHelper: saveHelper,
		postHelper: postHelper,
	}
}
