package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

type DeleteCacheRequest struct {
	CacheKey string `json:"cache_key" binding:"required"`
}

// Exposed method for interal service communication to delete cache
func (handlers *FeedHandlers) DeleteCache(c *gin.Context) {
	var dcr DeleteCacheRequest
	err := c.ShouldBindJSON(&dcr)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	cmd := handlers.cacheHelper.Del(dcr.CacheKey)
	if cmd.Err() != nil {
		utils.GeneralAPIInternalError(c, cmd.Err().Error())
		return
	}

	logging.Info("Successfully deleted cache key: ", dcr.CacheKey)

	utils.GenereateSuccessResponse(c, nil)
}
