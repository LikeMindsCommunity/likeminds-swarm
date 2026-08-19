package handlers

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/utils"
	"github.com/gin-gonic/gin"
)

type DeleteCacheRequest struct {
	CacheKey   string `json:"cache_key,omitempty"`
	KeyPattern string `json:"key_pattern,omitempty"`
}

// Exposed method for interal service communication to delete cache
func (handlers *FeedHandlers) DeleteCache(c *gin.Context) {
	var dcr DeleteCacheRequest
	err := c.ShouldBindJSON(&dcr)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	if dcr.KeyPattern != "" {
		cmd := handlers.cacheHelper.GetKeysFromPattern(dcr.KeyPattern)

		if cmd.Err() != nil {
			utils.GeneralAPIInternalError(c, cmd.Err().Error())
			return
		}

		handlers.cacheHelper.DelMultiple(cmd.Val())

	} else if dcr.CacheKey != "" {
		cmd := handlers.cacheHelper.Del(dcr.CacheKey)

		if cmd.Err() != nil {
			utils.GeneralAPIInternalError(c, cmd.Err().Error())
			return
		}

	} else {
		utils.GeneralAPIInternalError(c, "Send either cache key or key pattern!")
		return
	}

	logging.Info("Successfully deleted cache key: ", dcr.CacheKey)

	utils.GenerateSuccessResponse(c, nil)
}
