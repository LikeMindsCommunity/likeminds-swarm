package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Exposed Method to fetch the Universal Feed for a User
func (handlers *FeedHandlers) FetchUniversalFeed(c *gin.Context) {
	// fetch url params and headers
	headers := utils.GetHeaders(c)
	var universalFeedRequest requests.FetchUniversalFeedRequest

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	err := c.BindQuery(&universalFeedRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Parse topic Ids string array
	topicIds := parseStringArrayParam(universalFeedRequest.TopicIds)

	// fetch pagination query params
	page, _, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// pinned posts filter data
	pinnedPostFilterData := gin.H{
		"is_pinned":    true,
		"is_deleted":   false,
		"community_id": communityId,
		"$or": []gin.H{
			{
				"chatroom_id": gin.H{
					"$exists": false,
				},
			},
			{
				"chatroom_id": 0,
			},
		},
	}

	// unpinned posts filter data
	unpinnedPostFilterData := gin.H{
		"is_pinned":    false,
		"is_deleted":   false,
		"community_id": communityId,
		"$or": []gin.H{
			{
				"chatroom_id": gin.H{
					"$exists": false,
				},
			},
			{
				"chatroom_id": 0,
			},
		},
	}

	// Add topic id filter if topic_ids param exists
	if len(topicIds) > 0 {
		topicObjectIds := helpers.ConvertIdsToObjectIds(topicIds)

		pinnedPostFilterData["topic_ids"] = gin.H{
			"$in": topicObjectIds,
		}

		unpinnedPostFilterData["topic_ids"] = gin.H{
			"$in": topicObjectIds,
		}
	}

	// filter options
	postFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response := []requests.PostResponse{}

	if page == 1 {
		// pinned post filter options
		pinnedPostFilterOptions := addSortingOptions(map[string]interface{}{}, "created_at", OrderTypeDescending)

		// fetch pinned post using helper method
		pinnedPostResults, err := handlers.postHelper.FindPostHelper(pinnedPostFilterData,
			pinnedPostFilterOptions)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// parse pinned posts
		pinnedPostResponse := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
			handlers.saveHelper, handlers.topicHelper, pinnedPostResults, headers[utils.HeadersMemberId],
			universalFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
			apiRevampV1Check)

		response = append(response, pinnedPostResponse...)
	}

	// fetch unpinned post using helper method
	unpinnedPostResults, err := handlers.postHelper.FindPostHelper(unpinnedPostFilterData,
		postFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse unpinned posts
	unpinnedPostResponse := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, handlers.topicHelper, unpinnedPostResults, headers[utils.HeadersMemberId],
		universalFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check)

	response = append(response, unpinnedPostResponse...)

	finalResponse := parseFetchMultiplePostResponse(handlers.postHelper, response, -1)

	// reponse data
	finalParsedResponse := gin.H{
		"posts":       finalResponse.Posts,
		"success":     finalResponse.Success,
		"total_count": finalResponse.TotalCount,
	}

	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, communityId)

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}

// Internal Method to parse Explore feed for response
func parseExploreFeedResponse(chatroom_ids []int, post_counts map[int]int) requests.FetchExploreFeedResponse {
	var exploreResponse requests.FetchExploreFeedResponse

	exploreResponse.Success = true
	exploreResponse.ChatroomIDs = chatroom_ids
	exploreResponse.PostCounts = post_counts

	return exploreResponse
}

// Internal Method to parse ChatroomIds to int list
func parseChatroomIds(chatrooms []gin.H) []int {
	chatroomIds := []int{}
	for _, chatroom := range chatrooms {
		if chatroomId, ok := chatroom["chatroom_id"]; ok {
			chatroomIds = append(chatroomIds, int(chatroomId.(int32)))
		}
	}

	return chatroomIds
}

// Internal Method to get posts count in a Chatroom
func getPostCountInChatrooms(postHelper interfaces.PostHelper, chatrooms []int) map[int]int {
	postCountResponse := map[int]int{}
	postFilterData := []map[string]interface{}{}

	// Add match logic
	postFilterData = append(postFilterData, gin.H{
		"$match": gin.H{
			"is_deleted": false,
			"chatroom_id": gin.H{
				"$exists": true,
				"$in":     chatrooms,
			},
		},
	})

	// Add group logic
	postFilterData = append(postFilterData, gin.H{
		"$group": gin.H{
			"_id": "$chatroom_id",
			"post_count": gin.H{
				"$sum": 1,
			},
		},
	})

	// Add projection logic
	postFilterData = append(postFilterData, gin.H{
		"$project": gin.H{
			"_id":         0,
			"chatroom_id": "$_id",
			"post_count":  "$post_count",
		},
	})

	// fetch post using helper method
	postResults, err := postHelper.AggregatePostHelper(postFilterData)
	if err == nil {
		for _, chatroom := range postResults {
			chatroomId, ok1 := chatroom["chatroom_id"]
			postCount, ok2 := chatroom["post_count"]

			if ok1 && ok2 {
				postCountResponse[int(chatroomId.(int32))] = int(postCount.(int32))
			}
		}
	}

	return postCountResponse
}

// Internal Method to fetch Chatrooms ordered by recency of activity
func getChatroomsBasedOnRecentActivity(c *gin.Context, postHelper interfaces.PostHelper,
	communityId int, excludedChatroomIds []int, page int, pageSize int) []int {
	postFilterData := []map[string]interface{}{}

	// Add match logic
	postFilterData = append(postFilterData, gin.H{
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
	postFilterData = append(postFilterData, gin.H{
		"$group": gin.H{
			"_id": "$chatroom_id",
			"created_at": gin.H{
				"$max": "$created_at",
			},
		},
	})

	// Add sorting logic
	postFilterData = append(postFilterData, gin.H{
		"$sort": gin.H{
			"created_at": -1,
		},
	})

	// Add projection logic
	postFilterData = append(postFilterData, gin.H{
		"$project": gin.H{
			"_id":         0,
			"chatroom_id": "$_id",
		},
	})

	// Add pagination logic
	postFilterData = append(postFilterData, gin.H{
		"$skip": pageSize * (page - 1),
	})

	postFilterData = append(postFilterData, gin.H{
		"$limit": pageSize,
	})

	// fetch post using helper method
	postResults, err := postHelper.AggregatePostHelper(postFilterData)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return []int{}
	}

	chatroomIds := parseChatroomIds(postResults)

	return chatroomIds
}

// Internal Method to fetch Chatrooms ordered by count of messages
func getChatroomsBasedOnMostMessages(c *gin.Context, postHelper interfaces.PostHelper, communityId int,
	excludedChatroomIds []int, page int, pageSize int) []int {
	postFilterData := []map[string]interface{}{}

	// Add match logic
	postFilterData = append(postFilterData, gin.H{
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
	postFilterData = append(postFilterData, gin.H{
		"$group": gin.H{
			"_id": "$chatroom_id",
			"post_count": gin.H{
				"$sum": 1,
			},
		},
	})

	// Add sorting logic
	postFilterData = append(postFilterData, gin.H{
		"$sort": gin.H{
			"post_count": -1,
		},
	})

	// Add projection logic
	postFilterData = append(postFilterData, gin.H{
		"$project": gin.H{
			"_id":         0,
			"chatroom_id": "$_id",
		},
	})

	// Add pagination logic
	postFilterData = append(postFilterData, gin.H{
		"$skip": pageSize * (page - 1),
	})

	postFilterData = append(postFilterData, gin.H{
		"$limit": pageSize,
	})

	// fetch post using helper method
	postResults, err := postHelper.AggregatePostHelper(postFilterData)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return []int{}
	}

	chatroomIds := parseChatroomIds(postResults)

	return chatroomIds
}

// Exposed Method to fetch the Explore Feed
func (handlers *FeedHandlers) FetchExploreFeed(c *gin.Context) {
	// fetch url params and headers
	// headers := utils.GetHeaders(c)
	var exploreFeedRequest requests.FetchExploreFeedRequest

	err := c.BindQuery(&exploreFeedRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pagination query params
	page, pageSize, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// list of valid order types
	validOrderTypes := []int{constants.GroupOrderTypeNewest, constants.GroupOrderTypeRecentlyActive,
		constants.GroupOrderTypeMostMessages, constants.GroupOrderTypeMostParticipants}
	isValidOrderType := false

	// Validation of order types
	for _, orderType := range validOrderTypes {
		if orderType == exploreFeedRequest.OrderType {
			isValidOrderType = true
		}
	}

	if !isValidOrderType {
		utils.GeneralAPIValidationError(c, "Invalid order_type")
		return
	}

	chatroomIds := []int{}

	// Order by newest chatroom on top
	if (exploreFeedRequest.OrderType == constants.GroupOrderTypeNewest ||
		exploreFeedRequest.OrderType == constants.GroupOrderTypeMostParticipants) &&
		len(exploreFeedRequest.ChatroomIDs) > 0 {
		chatroomIds = parseIntArrayParam(exploreFeedRequest.ChatroomIDs)
	} else

	// Order by Recently active chatroom on top
	if exploreFeedRequest.OrderType == constants.GroupOrderTypeRecentlyActive {
		chatroomIds = getChatroomsBasedOnRecentActivity(c, handlers.postHelper, communityId,
			parseIntArrayParam(exploreFeedRequest.ExcludedChatroomIDs), page, pageSize)
	} else

	// Order by Most messaged chatroom on top
	if exploreFeedRequest.OrderType == constants.GroupOrderTypeMostMessages {
		chatroomIds = getChatroomsBasedOnMostMessages(c, handlers.postHelper, communityId,
			parseIntArrayParam(exploreFeedRequest.ExcludedChatroomIDs), page, pageSize)
	}

	postData := getPostCountInChatrooms(handlers.postHelper, chatroomIds)

	// return final response
	c.JSON(http.StatusOK, parseExploreFeedResponse(chatroomIds, postData))
}

// Exposed Method to fetch the Group Feed
func (handlers *FeedHandlers) FetchGroupFeed(c *gin.Context) {
	// fetch url params and headers
	headers := utils.GetHeaders(c)
	var groupFeedRequest requests.FetchGroupFeedRequest

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	err := c.BindQuery(&groupFeedRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Feedroom Id Validation
	if groupFeedRequest.FeedroomId == "" {
		utils.GeneralAPIValidationError(c, "send feedroom_id in params")
		return
	}

	// Conversation of feedroom Id from string to Int
	feedroomId, _ := strconv.Atoi(groupFeedRequest.FeedroomId)

	// Parse topic Ids string array
	topicIds := parseStringArrayParam(groupFeedRequest.TopicIds)

	// fetch pagination query params
	page, _, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// pinned posts filter data
	pinnedPostFilterData := gin.H{
		"is_pinned":    true,
		"is_deleted":   false,
		"community_id": communityId,
		"chatroom_id":  feedroomId,
	}

	// unpinned posts filter data
	unpinnedPostFilterData := gin.H{
		"is_pinned":    false,
		"is_deleted":   false,
		"community_id": communityId,
		"chatroom_id":  feedroomId,
	}

	// Add topic id filter if topic_ids param exists
	if len(topicIds) > 0 {
		topicObjectIds := helpers.ConvertIdsToObjectIds(topicIds)

		pinnedPostFilterData["topic_ids"] = gin.H{
			"$in": topicObjectIds,
		}

		unpinnedPostFilterData["topic_ids"] = gin.H{
			"$in": topicObjectIds,
		}
	}

	// filter options
	postFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response := []requests.PostResponse{}

	if page == 1 {
		// pinned post filter options
		pinnedPostFilterOptions := addSortingOptions(map[string]interface{}{}, "created_at", OrderTypeDescending)

		// fetch pinned post using helper method
		pinnedPostResults, err := handlers.postHelper.FindPostHelper(pinnedPostFilterData,
			pinnedPostFilterOptions)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// parse pinned posts
		pinnedPostResponse := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
			handlers.saveHelper, handlers.topicHelper, pinnedPostResults, headers[utils.HeadersMemberId],
			groupFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
			apiRevampV1Check)

		response = append(response, pinnedPostResponse...)
	}

	// fetch unpinned post using helper method
	unpinnedPostResults, err := handlers.postHelper.FindPostHelper(unpinnedPostFilterData,
		postFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse unpinned posts
	unpinnedPostResponse := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, handlers.topicHelper, unpinnedPostResults, headers[utils.HeadersMemberId],
		groupFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check)

	response = append(response, unpinnedPostResponse...)

	finalResponse := parseFetchMultiplePostResponse(handlers.postHelper, response, -1)

	// reponse data
	finalParsedResponse := gin.H{
		"posts":       finalResponse.Posts,
		"success":     finalResponse.Success,
		"total_count": finalResponse.TotalCount,
	}

	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, communityId)

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}
