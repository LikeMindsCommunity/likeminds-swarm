package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func (handlers *FeedHandlers) FetchUniversalFeed(c *gin.Context) {
	// fetch url params and headers
	headers := utils.GetHeaders(c)
	param_is_cm := c.Query("user_is_cm")
	is_cm := false

	if param_is_cm == "true" {
		is_cm = true
	}

	// fetch pagination query params
	page, _, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// pinned posts filter data
	pinned_post_filter_data := gin.H{
		"is_pinned":    true,
		"is_deleted":   false,
		"community_id": community_id,
	}

	// unpinned posts filter data
	unpinned_post_filter_data := gin.H{
		"is_pinned":    false,
		"is_deleted":   false,
		"community_id": community_id,
	}

	// filter options
	post_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response := []requests.PostResponse{}

	if page == 1 {
		// fetch pinned post using helper method
		pinned_post_results, err := handlers.postHelper.FindPostHelper(pinned_post_filter_data, gin.H{})
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// parse pinned posts
		pinned_post_response := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper,
			pinned_post_results, headers[utils.HeadersMemberId], is_cm)

		response = append(response, pinned_post_response...)
	}

	// fetch unpinned post using helper method
	unpinned_post_results, err := handlers.postHelper.FindPostHelper(unpinned_post_filter_data, post_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse unpinned posts
	unpinned_post_response := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper,
		unpinned_post_results, headers[utils.HeadersMemberId], is_cm)

	response = append(response, unpinned_post_response...)

	// return final response
	c.JSON(http.StatusOK, parseFetchMultiplePostResponse(handlers.postHelper, response, -1))
}
