package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
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

func parseExploreFeedResponse(chatroom_ids []int, post_counts map[int]int) requests.FetchExploreFeedResponse {
	var exploreResponse requests.FetchExploreFeedResponse

	exploreResponse.Success = true
	exploreResponse.ChatroomIDs = chatroom_ids
	exploreResponse.PostCounts = post_counts

	return exploreResponse
}

func parseChatroomIds(chatrooms []gin.H) []int {
	chatroom_ids := []int{}
	for _, chatroom := range chatrooms {
		if chatroom_id, ok := chatroom["chatroom_id"]; ok {
			chatroom_ids = append(chatroom_ids, int(chatroom_id.(int32)))
		}
	}

	return chatroom_ids
}

func getPostCountInChatrooms(postHelper interfaces.PostHelper, chatrooms []int) map[int]int {
	post_count_response := map[int]int{}
	post_filter_data := []map[string]interface{}{}

	// Add match logic
	post_filter_data = append(post_filter_data, gin.H{
		"$match": gin.H{
			"is_deleted": false,
			"chatroom_id": gin.H{
				"$exists": true,
				"$nin":    []int{0},
			},
		},
	})

	// Add group logic
	post_filter_data = append(post_filter_data, gin.H{
		"$group": gin.H{
			"_id": "$chatroom_id",
			"post_count": gin.H{
				"$sum": 1,
			},
		},
	})

	// Add projection logic
	post_filter_data = append(post_filter_data, gin.H{
		"$project": gin.H{
			"_id":         0,
			"chatroom_id": "$_id",
			"post_count":  "$post_count",
		},
	})

	// fetch post using helper method
	post_results, err := postHelper.AggregatePostHelper(post_filter_data)
	if err == nil {
		for _, chatroom := range post_results {
			chatroom_id, ok1 := chatroom["chatroom_id"]
			post_count, ok2 := chatroom["post_count"]

			if ok1 && ok2 {
				post_count_response[int(chatroom_id.(int32))] = int(post_count.(int32))
			}
		}
	}

	return post_count_response
}

func getChatroomsBasedOnRecentActivity(c *gin.Context, postHelper interfaces.PostHelper, communityId int, excludedChatroomIds []int, page int, page_size int) []int {
	post_filter_data := []map[string]interface{}{}

	// Add match logic
	post_filter_data = append(post_filter_data, gin.H{
		"$match": gin.H{
			"is_deleted":   false,
			"community_id": communityId,
			"chatroom_id": gin.H{
				"$exists": true,
				"$nin":    append(excludedChatroomIds, 0),
			},
		},
	})

	// Add group logic
	post_filter_data = append(post_filter_data, gin.H{
		"$group": gin.H{
			"_id": "$chatroom_id",
			"created_at": gin.H{
				"$max": "$created_at",
			},
		},
	})

	// Add sorting logic
	post_filter_data = append(post_filter_data, gin.H{
		"$sort": gin.H{
			"created_at": -1,
		},
	})

	// Add projection logic
	post_filter_data = append(post_filter_data, gin.H{
		"$project": gin.H{
			"_id":         0,
			"chatroom_id": "$_id",
		},
	})

	// Add pagination logic
	post_filter_data = append(post_filter_data, gin.H{
		"$skip": page_size * (page - 1),
	})

	post_filter_data = append(post_filter_data, gin.H{
		"$limit": page_size,
	})

	// fetch post using helper method
	post_results, err := postHelper.AggregatePostHelper(post_filter_data)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return []int{}
	}

	chatroom_ids := parseChatroomIds(post_results)

	return chatroom_ids
}

func getChatroomsBasedOnMostMessages(c *gin.Context, postHelper interfaces.PostHelper, communityId int, excludedChatroomIds []int, page int, page_size int) []int {
	post_filter_data := []map[string]interface{}{}

	// Add match logic
	post_filter_data = append(post_filter_data, gin.H{
		"$match": gin.H{
			"is_deleted":   false,
			"community_id": communityId,
			"chatroom_id": gin.H{
				"$exists": true,
				"$nin":    append(excludedChatroomIds, 0),
			},
		},
	})

	// Add group logic
	post_filter_data = append(post_filter_data, gin.H{
		"$group": gin.H{
			"_id": "$chatroom_id",
			"post_count": gin.H{
				"$sum": 1,
			},
		},
	})

	// Add sorting logic
	post_filter_data = append(post_filter_data, gin.H{
		"$sort": gin.H{
			"post_count": -1,
		},
	})

	// Add projection logic
	post_filter_data = append(post_filter_data, gin.H{
		"$project": gin.H{
			"_id":         0,
			"chatroom_id": "$_id",
		},
	})

	// Add pagination logic
	post_filter_data = append(post_filter_data, gin.H{
		"$skip": page_size * (page - 1),
	})

	post_filter_data = append(post_filter_data, gin.H{
		"$limit": page_size,
	})

	// fetch post using helper method
	post_results, err := postHelper.AggregatePostHelper(post_filter_data)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return []int{}
	}

	chatroom_ids := parseChatroomIds(post_results)

	return chatroom_ids
}

func (handlers *FeedHandlers) FetchExploreFeed(c *gin.Context) {
	// fetch url params and headers
	// headers := utils.GetHeaders(c)
	var exploreFeedRequest requests.FetchExploreFeedRequest

	err := c.ShouldBindQuery(&exploreFeedRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pagination query params
	page, page_size, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// list of valid order types
	validOrderTypes := []int{constants.GroupOrderTypeNewest, constants.GroupOrderTypeRecentlyActive, constants.GroupOrderTypeMostMessages, constants.GroupOrderTypeMostParticipants}
	isValidOrderType := false

	// Validation of order types
	for _, order_type := range validOrderTypes {
		if order_type == exploreFeedRequest.OrderType {
			isValidOrderType = true
		}
	}

	if !isValidOrderType {
		utils.GeneralAPIValidationError(c, "Invalid order_type")
		return
	}

	chatroom_ids := []int{}

	// Order by newest chatroom on top
	if (exploreFeedRequest.OrderType == constants.GroupOrderTypeNewest || exploreFeedRequest.OrderType == constants.GroupOrderTypeMostMessages) && len(exploreFeedRequest.ChatroomIDs) > 0 {
		chatroom_ids = exploreFeedRequest.ChatroomIDs
	} else

	// Order by Recently active chatroom on top
	if exploreFeedRequest.OrderType == constants.GroupOrderTypeRecentlyActive {
		chatroom_ids = getChatroomsBasedOnRecentActivity(c, handlers.postHelper, community_id, exploreFeedRequest.ExcludedChatroomIDs, page, page_size)
	} else

	// Order by Most messaged chatroom on top
	if exploreFeedRequest.OrderType == constants.GroupOrderTypeMostMessages {
		chatroom_ids = getChatroomsBasedOnMostMessages(c, handlers.postHelper, community_id, exploreFeedRequest.ExcludedChatroomIDs, page, page_size)
	}

	postData := getPostCountInChatrooms(handlers.postHelper, chatroom_ids)

	// return final response
	c.JSON(http.StatusOK, parseExploreFeedResponse(chatroom_ids, postData))
}
