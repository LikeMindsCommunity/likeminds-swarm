package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to get Universal Feed Post Filter
func getUniversalFeedPostFilter(communityId int, userId string, excludedUserIds []string, isCm bool) gin.H {
	universalPostFilter := gin.H{
		"is_deleted":   false,
		"community_id": communityId,
		"user_id": gin.H{
			"$nin": excludedUserIds,
		},
		"$and": []gin.H{
			{
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
			},
			{
				"$or": []gin.H{
					{
						"visibility": gin.H{
							"$exists": false,
						},
					},
					{
						"visibility": enums.PublicVisibility,
					},
				},
			},
		},
	}

	// Add filter for hidden posts
	if !isCm {
		universalPostFilter["$nor"] = []gin.H{
			{
				"is_hidden": true,
				"user_id": gin.H{
					"$ne": userId,
				},
			},
		}
	}

	return universalPostFilter
}

// Exposed Method to fetch the Universal Feed for a User
func (handlers *FeedHandlers) FetchUniversalFeed(c *gin.Context) {

	defer utils.Timer("------------------FetchUniversalFeed------------------")()

	headers := utils.GetHeaders(c)
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]

	userId := headers[utils.HeadersMemberId]
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	var universalFeedRequest requests.FetchUniversalFeedRequest
	err := c.BindQuery(&universalFeedRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUserParams := LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             universalFeedRequest.IsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// Parse topic Ids string array
	topicIds := utils.ParseStringArrayParam(universalFeedRequest.TopicIds)

	// Parse widget ids string array
	widgetIds := utils.ParseStringArrayParam(universalFeedRequest.WidgetIds)

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// posts filter data
	postFilterData := getUniversalFeedPostFilter(communityId, userId, excludedUserIds, loggedInUserParams.IsCm)

	// Add topic id filter if topic_ids param exists
	if len(topicIds) > 0 {
		postObjectIdsList, err := getPostIdsBasedOnTopicsFilter(handlers, topicIds)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		if len(postObjectIdsList) == 0 {
			finalParsedResponse := gin.H{
				"posts":             []responses.PostResponse{},
				"topics":            map[string]responses.TopicResponse{},
				"widgets":           map[string]requests.WidgetResponse{},
				"reposted_posts":    map[string]responses.PostResponse{},
				"filtered_comments": map[string]responses.CommentWithParentResponse{},
			}

			// return final response
			utils.GenerateSuccessResponse(c, finalParsedResponse)
			return
		}

		postFilterData["_id"] = gin.H{
			"$in": postObjectIdsList,
		}
	}

	// Add widget id filter if widget_ids param exists
	if len(widgetIds) > 0 {
		widgetObjectIds := helpers.ConvertIdsToObjectIds(widgetIds)

		widgetQuery := gin.H{
			"$elemMatch": gin.H{
				"attachment_meta.entity_id": gin.H{
					"$in": widgetObjectIds,
				},
			},
		}

		postFilterData["attachments"] = widgetQuery
	}

	// filter options
	postFilterOptions, err := generatePageFilterOptions(c, "is_pinned", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Add post filter options
	postFilterOptions = addSortingOptions(postFilterOptions, "created_at", OrderTypeDescending)

	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, postFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse posts
	parsedPosts := parseMultiplePostResponse(handlers, postResults, headers[utils.HeadersMemberId], universalFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, memberRole)

	// parse posts for final response (topics, widgets, comments, etc)
	finalParsedResponse := parsePostsAndGenerateFinalResponse(handlers, &loggedInUserParams, parsedPosts)

	// return final response
	utils.GenerateSuccessResponse(c, finalParsedResponse)
}

// Internal method to parse posts for topics, reposted posts, widgets, comments, etc and generate final response
func parsePostsAndGenerateFinalResponse(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, parsedPosts []responses.PostResponse,
) gin.H {

	defer utils.Timer("parsePostsAndGenerateFinalResponse")()

	// final reponse data
	finalParsedResponse := gin.H{
		"posts": parsedPosts,
	}

	// get topic data from posts
	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, loggedInUser.CommunityId)

	// get reposted posts data
	finalParsedResponse["reposted_posts"] = getOriginalPostForReposts(handlers, finalParsedResponse, loggedInUser.CommunityId, loggedInUser.UserId, loggedInUser.IsCm, loggedInUser.VersionCode, loggedInUser.PlatformCode, loggedInUser.ApiRevampCheckV1)

	// get widget data from feed response data
	finalParsedResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalParsedResponse, loggedInUser.CommunityId, loggedInUser.IsCm, loggedInUser.UserId)

	// filter top comments against posts based on community config and update final response
	updatedPosts, filteredComments := filterTopCommentsForPosts(handlers, parsedPosts, *loggedInUser)
	finalParsedResponse["filtered_comments"] = filteredComments
	finalParsedResponse["posts"] = updatedPosts

	return finalParsedResponse
}

// Internal Method to filter top comments against Posts based on community config
func filterTopCommentsForPosts(handlers *FeedHandlers, parsedPosts []responses.PostResponse, LoggedInUser LoggedInUserParams,
) ([]responses.PostResponse, map[string]responses.CommentWithParentResponse) {

	defer utils.Timer("filterTopCommentsForPosts")()

	// Get community configurations
	universalFeedConfig := externalHelpers.GetUniversalFeedConfigurationsData(handlers.cacheHelper, LoggedInUser.UserId, LoggedInUser.CommunityId)

	var commentSortOrderVal int
	if universalFeedConfig.CommentSortOrder == enums.DescendingSortOrder {
		commentSortOrderVal = -1
	} else {
		commentSortOrderVal = 1
	}

	filteredComments := map[string]responses.CommentWithParentResponse{}

	if universalFeedConfig.CommentSortOn == enums.UniversalFeedTopLikedComments {

		var updatedPostsWithComments []responses.PostResponse
		var err error

		updatedPostsWithComments, filteredComments, err = getTopCommentsAgainstPostsSortOnLikes(handlers, parsedPosts,
			LoggedInUser.UserId, LoggedInUser.IsCm, LoggedInUser.CommunityId, commentSortOrderVal, universalFeedConfig.CommentCount,
			LoggedInUser.VersionCode, LoggedInUser.PlatformCode, LoggedInUser.ApiRevampCheckV1, LoggedInUser.MemberRole)
		if err != nil {
			logging.Error(constants.FetchingTopCommentsErrorMessage, err)
			return parsedPosts, filteredComments
		}

		if len(updatedPostsWithComments) > 0 {
			parsedPosts = updatedPostsWithComments
		}
	}

	return parsedPosts, filteredComments
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
		chatroomIds = utils.ParseIntArrayParam(exploreFeedRequest.ChatroomIDs)
	} else

	// Order by Recently active chatroom on top
	if exploreFeedRequest.OrderType == constants.GroupOrderTypeRecentlyActive {
		chatroomIds = getChatroomsBasedOnRecentActivity(c, handlers.postHelper, communityId,
			utils.ParseIntArrayParam(exploreFeedRequest.ExcludedChatroomIDs), page, pageSize)
	} else

	// Order by Most messaged chatroom on top
	if exploreFeedRequest.OrderType == constants.GroupOrderTypeMostMessages {
		chatroomIds = getChatroomsBasedOnMostMessages(c, handlers.postHelper, communityId,
			utils.ParseIntArrayParam(exploreFeedRequest.ExcludedChatroomIDs), page, pageSize)
	}

	postData := getPostCountInChatrooms(handlers.postHelper, chatroomIds)

	// return final response
	c.JSON(http.StatusOK, parseExploreFeedResponse(chatroomIds, postData))
}

// Internal Method to get Group Feed Post Filter
func getGroupFeedPostFilter(communityId int, userId string, feedroomId int, isCm bool) gin.H {
	groupFeedPostFilter := gin.H{
		"is_deleted":   false,
		"community_id": communityId,
		"chatroom_id":  feedroomId,
		"$and": []gin.H{
			{
				"$or": []gin.H{
					{
						"visibility": gin.H{
							"$exists": false,
						},
					},
					{
						"visibility": enums.PublicVisibility,
					},
				},
			},
		},
	}

	// Add filter for hidden posts
	if !isCm {
		groupFeedPostFilter["$nor"] = []gin.H{
			{
				"is_hidden": true,
				"user_id": gin.H{
					"$ne": userId,
				},
			},
		}
	}

	return groupFeedPostFilter
}

// Exposed Method to fetch the Group Feed
func (handlers *FeedHandlers) FetchGroupFeed(c *gin.Context) {

	// fetch url params and headers
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	var groupFeedRequest requests.FetchGroupFeedRequest
	err := c.BindQuery(&groupFeedRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUser := LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             groupFeedRequest.IsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// Feedroom Id Validation
	if groupFeedRequest.FeedroomId == "" {
		utils.GeneralAPIValidationError(c, "send feedroom_id in params")
		return
	}

	// Conversation of feedroom Id from string to Int
	feedroomId, _ := strconv.Atoi(groupFeedRequest.FeedroomId)

	// Parse topic Ids string array
	topicIds := utils.ParseStringArrayParam(groupFeedRequest.TopicIds)

	// unpinned posts filter data
	groupFeedPostFilter := getGroupFeedPostFilter(communityId, userId, feedroomId, loggedInUser.IsCm)

	// Add topic id filter if topic_ids param exists
	if len(topicIds) > 0 {
		topicObjectIds := helpers.ConvertIdsToObjectIds(topicIds)

		groupFeedPostFilter["topic_ids"] = gin.H{
			"$in": topicObjectIds,
		}
	}

	// filter options
	postFilterOptions, err := generatePageFilterOptions(c, "is_pinned", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Add post filter options
	postFilterOptions = addSortingOptions(postFilterOptions, "created_at", OrderTypeDescending)

	// fetch posts using helper method
	postResults, err := handlers.postHelper.FindPostHelper(groupFeedPostFilter,
		postFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse posts
	parsedPosts := parseMultiplePostResponse(handlers, postResults, headers[utils.HeadersMemberId], groupFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, utils.DefaultRole)

	// parse posts for final response (topics, widgets, comments, etc)
	finalParsedResponse := parsePostsAndGenerateFinalResponse(handlers, &loggedInUser, parsedPosts)

	// return final response
	utils.GenerateSuccessResponse(c, finalParsedResponse)
}

// WarmupCommunityUniversaFeedCache | push community universal feed first page to cache
func (handlers *FeedHandlers) WarmupCommunityUniversaFeedCache(communityID int) {
	handlers.deleteCommunityUniversalFeedCacheData(communityID)

	// add create logic
}

func (handlers *FeedHandlers) deleteCommunityUniversalFeedCacheData(communityID int) {
	cacheCommunityUniversalFeedPostsKey := fmt.Sprintf("community_%d_universal_feed_posts", communityID)

	cacheCommunityUniversleFeedPostIDsString := handlers.cacheHelper.Get(cacheCommunityUniversalFeedPostsKey)
	cacheCommunityUniversleFeedPostIDs := []string{cacheCommunityUniversleFeedPostIDsString.Val()}

	cachePostKeys := []string{}
	for _, cacheCommunityUniversleFeedPostID := range cacheCommunityUniversleFeedPostIDs {
		cachePostKey := fmt.Sprintf("post_%s", cacheCommunityUniversleFeedPostID)
		cachePostKeys = append(cachePostKeys, cachePostKey)
	}

	handlers.cacheHelper.DelMultiple(cachePostKeys)
	handlers.cacheHelper.Del(cacheCommunityUniversalFeedPostsKey)
}

// Internal Method to get Collection Feed Items Filter
func getConnectionFeedItemFilter(communityId int, userId string) gin.H {
	return gin.H{
		"user_id":      userId,
		"community_id": communityId,
	}
}

// Internal Method to get Connection Feed Post Filter
func getConnectionFeedPostFilter(userId string, communityId int, postIds []primitive.ObjectID, isCm bool) gin.H {
	connectionFeedPostFilter := gin.H{
		"is_deleted":   false,
		"community_id": communityId,
		"_id": gin.H{
			"$in": postIds,
		},
	}

	// Add filter for hidden posts
	if !isCm {
		connectionFeedPostFilter["$nor"] = []gin.H{
			{
				"is_hidden": true,
				"user_id": gin.H{
					"$ne": userId,
				},
			},
		}
	}

	return connectionFeedPostFilter
}

// Exposed Method to Fetch Connection Feed for a User
func (handlers *FeedHandlers) FetchConnectionFeed(c *gin.Context) {
	// fetch headers
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var connectionFeedRequest requests.FetchConnectionFeedRequest
	if err := c.BindQuery(&connectionFeedRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUser := LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             connectionFeedRequest.IsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// get feed buffer data for the user
	userConnectionFeedData, cacheKeyExists := getConnectionFeedBufferDataFromCache(handlers, userId, communityId)
	if !cacheKeyExists {
		// warm up connection list
		updateConnectionFeedBuffer(handlers, userId, communityId, "", false)
	} else {
		// convert post ids to object Ids
		postIds := []string{}
		for postId := range userConnectionFeedData {
			postIds = append(postIds, postId)
		}

		updatedPostIds := helpers.ConvertIdsToObjectIds(postIds)

		// Create record in DB and update data in cache
		for _, postId := range updatedPostIds {
			_, err := handlers.connectionFeedHelper.CreateConnectionFeedHelper(postId, userId, communityId)
			if err == nil {
				updateConnectionFeedBuffer(handlers, userId, communityId, postId.Hex(), false)
			}
		}
	}

	// connection feed filter data
	connectionFeedFilterData := getConnectionFeedItemFilter(communityId, userId)

	// fetch connection feed data using helper method
	connectionFeedResults, err := handlers.connectionFeedHelper.FindConnectionFeedHelper(connectionFeedFilterData, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	connectionFeedPostIds := []primitive.ObjectID{}

	for _, connectionFeedItem := range connectionFeedResults {
		connectionFeedPostIds = append(connectionFeedPostIds, connectionFeedItem.PostId)
	}

	// connection feed post filter data
	connectionFeedPostFilterData := getConnectionFeedPostFilter(userId, communityId, connectionFeedPostIds, loggedInUser.IsCm)

	// connection feed post filter options
	connectionFeedPostFilterOptions := addSortingOptions(map[string]interface{}{}, "created_at", OrderTypeDescending)

	// fetch connection feed post using helper method
	connectionFeedPostResults, err := handlers.postHelper.FindPostHelper(connectionFeedPostFilterData,
		connectionFeedPostFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// parse connection feed posts
	parsedPosts := parseMultiplePostResponse(handlers, connectionFeedPostResults, headers[utils.HeadersMemberId], connectionFeedRequest.IsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, utils.DefaultRole)

	// parse posts for final response (topics, widgets, comments, etc)
	finalParsedResponse := parsePostsAndGenerateFinalResponse(handlers, &loggedInUser, parsedPosts)

	// return final response
	utils.GenerateSuccessResponse(c, finalParsedResponse)
}

func getPostLikesCountAgainstUserQuery(userId string) []map[string]interface{} {
	postLikesFilterData := []map[string]interface{}{}

	// Add filter logic
	postLikesFilterData = append(postLikesFilterData, gin.H{
		"$match": gin.H{
			"is_deleted": false,
		},
	})

	// Add lookup logic
	postLikesFilterData = append(postLikesFilterData, gin.H{
		"$lookup": gin.H{
			"from": "like",
			"let": gin.H{
				"postId": "$_id",
			},
			"pipeline": []gin.H{
				{
					"$match": gin.H{
						"$expr": gin.H{
							"$eq": []string{"$entity_id", "$$postId"},
						},
						"is_deleted": false,
						"liked_by":   userId,
					},
				},
			},
			"as": "likes_data",
		},
	})

	// Add group logic
	postLikesFilterData = append(postLikesFilterData, gin.H{
		"$group": gin.H{
			"_id": "",
			"total_likes_count": gin.H{
				"$sum": gin.H{
					"$size": "$likes_data",
				},
			},
		},
	})

	return postLikesFilterData

}

func getPostLikesCountAgainstUser(postHelper interfaces.PostHelper, userId string) (int32, error) {
	var userPostLikesCount int32

	userPostLikesFilterData := getPostLikesCountAgainstUserQuery(userId)

	userPostLikesCountData, err := postHelper.AggregatePostHelper(userPostLikesFilterData)
	if err != nil {
		return 0, err
	}

	if len(userPostLikesCountData) > 0 {
		if userPostLikesCountMap, ok := userPostLikesCountData[0]["total_likes_count"]; ok {
			userPostLikesCount = userPostLikesCountMap.(int32)
		}
	}

	return userPostLikesCount, nil
}

// Exposed Method to fetch user feed meta
func (handlers *FeedHandlers) FetchUserFeedMeta(c *gin.Context) {
	// fetch url params and headers
	userId := c.Param("user_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// post filter data
	postFilterData := gin.H{
		"user_id":      userId,
		"is_deleted":   false,
		"community_id": communityId,
	}

	// fetch posts count using helper method
	postsCount, err := handlers.postHelper.CountPostHelper(postFilterData)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// comment filter data
	commentFilterData := gin.H{
		"user_id":      userId,
		"is_deleted":   false,
		"community_id": communityId,
		"level":        0,
	}

	commentsCount, err := handlers.commentHelper.CountCommentHelper(commentFilterData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Get user post likes count data
	userPostLikesCount, err := getPostLikesCountAgainstUser(handlers.postHelper, userId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Get user pending posts count
	pendingPostCountFilter := gin.H{
		"status": gin.H{
			"$in": []string{
				enums.UnderReview,
				enums.Rejected,
			},
		},
		"is_deleted": false,
		"user_id":    userId,
	}
	userPendingPostsCount, err := handlers.pendingPostHelper.CountPendingPostHelper(pendingPostCountFilter)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// response data
	finalResponse := gin.H{
		"posts_count":         postsCount,
		"comments_count":      commentsCount,
		"posts_like_count":    userPostLikesCount,
		"pending_posts_count": userPendingPostsCount,
	}

	// return final response
	utils.GenerateSuccessResponse(c, finalResponse)
}
