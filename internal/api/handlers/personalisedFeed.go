package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Exposed Method to Add a New Poll Option
func (handlers *FeedHandlers) RecomputePersonalisedFeed(c *gin.Context) {
	// fetch headers and url params
	// headers := utils.GetHeaders(c)

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	utils.GenerateSuccessResponse(c, gin.H{})
}
