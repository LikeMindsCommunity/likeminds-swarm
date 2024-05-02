package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getPostRepostCount(widgetHelper interfaces.WidgetHelper, post entities.Post) int32 {
	var postRepostCount int32 = 0

	postRepostWidgetData := getRepostWidgetDataFromPost(post)
	if postRepostWidgetData.AttachmentType == enums.RepostWidget {
		repostWidgetID := postRepostWidgetData.AttachmentMeta.EntityID

		widgetFilter := gin.H{
			"_id": repostWidgetID,
		}
		repostWidgets, err := widgetHelper.FindWidgetHelper(widgetFilter, gin.H{})
		if err != nil {
			return postRepostCount
		}

		if len(repostWidgets) <= 0 {
			return postRepostCount
		}

		return repostWidgets[0].MetaData["repost_count"].(int32)
	}

	return postRepostCount
}

func getIsRepostedByUser(widgetHelper interfaces.WidgetHelper, userID string, post entities.Post) bool {
	originalPostRepostWidgetData := getRepostWidgetDataFromPost(post)
	if originalPostRepostWidgetData.AttachmentType == enums.RepostWidget {
		//get repost widget id, update repost widget data
		repostWidgetID := originalPostRepostWidgetData.AttachmentMeta.EntityID

		widgetFilter := gin.H{
			"_id": repostWidgetID,
		}
		repostWidgets, err := widgetHelper.FindWidgetHelper(widgetFilter, gin.H{})
		if err != nil {
			return false
		}

		repostWidgetData := repostWidgets[0]
		repostWidgetMetadata := repostWidgetData.MetaData
		repostWidgetMetadataReposts := repostWidgetMetadata["reposts"]
		repostWidgetMetadataRepostsMap, ok := repostWidgetMetadataReposts.(map[string]interface{})
		if !ok {
			return false
		}
		if _, ok := repostWidgetMetadataRepostsMap[userID]; ok {
			return true
		}
	}
	return false
}

func validateUserAndPostForRepost(handlers *FeedHandlers, userID string, originalPostID string, communityId int) error {
	postFilterData := gin.H{
		"_id": originalPostID,
	}
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
	if (err != nil) || (len(postResults) <= 0) {
		return fmt.Errorf("original post not found for repost")
	}

	if getIsRepostedByUser(handlers.widgetHelper, userID, postResults[0]) {
		return fmt.Errorf("can not repost one post multiple times")
	}

	if postResults[0].IsDeleted {
		return fmt.Errorf("can not repost a deleted post")
	}

	if communityId != postResults[0].CommunityId {
		return fmt.Errorf("invalid post")
	}

	if postResults[0].IsRepost {
		return fmt.Errorf("can not repost a repost")
	}

	return nil
}

// Internal Method to parse response for fetch multiple posts api
func parseFetchMultiplePostResponse(posts []responses.PostResponse, posts_count int64,
) responses.FetchUserMultiplePostResponse {

	response := responses.FetchUserMultiplePostResponse{}

	response.Success = true
	response.Posts = posts

	if posts_count > 0 {
		response.TotalCount = int(posts_count)
	}

	return response
}

// Internal Method to parse topics response
func fetchAndParseTopicsForResponse(topicHelper interfaces.TopicHelper, topicIds []primitive.ObjectID,
	communityId int) (map[string]responses.TopicResponse, error) {

	// Fetch topics using topic Ids
	topics, err := fetchTopicsByIDs(topicHelper, topicIds, communityId, false)
	if err != nil {
		return nil, err
	}

	topicsResponse := map[string]responses.TopicResponse{}

	// Parse all fetched topics Data
	for _, topic := range topics {
		topicsResponse[topic.ID.Hex()] = parseTopicResponse(&topic)
	}

	return topicsResponse, nil
}

// Internal Method to parse widgets response
func parseWidgetsResponse(handlers *FeedHandlers, widgetIds []primitive.ObjectID, communityId int, userIsCM bool,
	userId string) (map[string]requests.WidgetResponse, error) {

	// Fetch widgets using widget Ids
	widgets, err := fetchWidgetsByIDs(handlers.widgetHelper, widgetIds, communityId)
	if err != nil {
		return nil, err
	}

	widgetsResponse := map[string]requests.WidgetResponse{}

	// Parse all fetched widgets Data
	for _, widget := range widgets {
		widgetsResponse[widget.ID.Hex()] = parseWidgetResponse(handlers, &widget, communityId, userIsCM, userId)
	}

	return widgetsResponse, nil
}

// Internal Method to parse topic_ids from posts
func getTopicIdsFromPosts(response interface{}) []primitive.ObjectID {
	uniqueTopicIds := []primitive.ObjectID{}
	tempTopicIds := map[primitive.ObjectID]bool{}

	if post, ok := response.(gin.H)["post"]; ok {

		switch post := post.(type) {
		case responses.PostWithRepliesResponse:
			for _, topicId := range post.Topics {
				if _, exists := tempTopicIds[topicId]; !exists {
					tempTopicIds[topicId] = true
				}
			}
		}
	}

	if posts, ok := response.(gin.H)["posts"]; ok {

		switch posts := posts.(type) {
		case []responses.PostResponse:
			for _, post := range posts {
				for _, topicId := range post.Topics {
					if _, exists := tempTopicIds[topicId]; !exists {
						tempTopicIds[topicId] = true
					}
				}
			}
		case map[string]responses.PostResponse:
			for _, post := range posts {
				for _, topicId := range post.Topics {
					if _, exists := tempTopicIds[topicId]; !exists {
						tempTopicIds[topicId] = true
					}
				}
			}
		}
	}

	for key := range tempTopicIds {
		uniqueTopicIds = append(uniqueTopicIds, key)
	}

	return uniqueTopicIds
}

// Internal method to parse widget_ids from topics
func getWidgetIdsFromTopics(response interface{}) []primitive.ObjectID {

	uniqueWidgetIds, widgetsMap := []primitive.ObjectID{}, map[string]bool{}

	if topic, ok := response.(gin.H)["topic"]; ok {
		widgetsMap = typeAssertAndFetchWidgetIdsFromTopics(topic, widgetsMap)
	}

	if topics, ok := response.(gin.H)["topics"]; ok {
		widgetsMap = typeAssertAndFetchWidgetIdsFromTopics(topics, widgetsMap)
	}

	if topics, ok := response.(gin.H)["child_topics"]; ok {
		widgetsMap = typeAssertAndFetchWidgetIdsFromTopics(topics, widgetsMap)
	}

	// convert map to array
	for key := range widgetsMap {
		objectId, _ := primitive.ObjectIDFromHex(key)
		uniqueWidgetIds = append(uniqueWidgetIds, objectId)
	}

	return uniqueWidgetIds
}

func getWidgetIdsFromComments(response interface{}) []primitive.ObjectID {

	uniqueWidgetIds := []primitive.ObjectID{}
	tempWidgetIds := map[primitive.ObjectID]bool{}

	if comment, ok := response.(gin.H)["comment"]; ok {
		tempWidgetIds = typeAssertAndFetchWidgetIdsFromComments(comment, tempWidgetIds)
	}

	if comments, ok := response.(gin.H)["comments"]; ok {
		tempWidgetIds = typeAssertAndFetchWidgetIdsFromComments(comments, tempWidgetIds)
	}

	if comments, ok := response.(gin.H)["replies"]; ok {
		tempWidgetIds = typeAssertAndFetchWidgetIdsFromComments(comments, tempWidgetIds)
	}

	for key := range tempWidgetIds {
		uniqueWidgetIds = append(uniqueWidgetIds, key)
	}

	return uniqueWidgetIds
}

func typeAssertAndFetchWidgetIdsFromComments(comments interface{}, widgetMap map[primitive.ObjectID]bool) map[primitive.ObjectID]bool {

	typeOf := fmt.Sprintf("%T", comments)
	print(typeOf)
	switch comments := comments.(type) {
	case responses.CommentResponse:
		widgetIds := getWidgetIdsFromAttachments(comments.Attachments)

		for _, widgetId := range widgetIds {

			if _, exists := widgetMap[widgetId]; !exists {
				widgetMap[widgetId] = true
			}
		}

	case responses.CommentWithParentResponse:
		tempAttachments := comments.Attachments

		if comments.ParentComment != nil {
			tempAttachments = append(tempAttachments, comments.ParentComment.Attachments...)
		}

		widgetIds := getWidgetIdsFromAttachments(tempAttachments)

		for _, widgetId := range widgetIds {
			if _, exists := widgetMap[widgetId]; !exists {
				widgetMap[widgetId] = true
			}
		}

	case []responses.CommentResponse:
		for _, comment := range comments {
			widgetIds := getWidgetIdsFromAttachments(comment.Attachments)

			for _, widgetId := range widgetIds {
				if _, exists := widgetMap[widgetId]; !exists {
					widgetMap[widgetId] = true
				}
			}
		}

	case map[string]responses.CommentResponse:
		for _, comment := range comments {
			widgetIds := getWidgetIdsFromAttachments(comment.Attachments)

			for _, widgetId := range widgetIds {
				if _, exists := widgetMap[widgetId]; !exists {
					widgetMap[widgetId] = true
				}
			}
		}

	case map[string][]responses.CommentResponse:
		for _, comment := range comments {
			for _, comment := range comment {
				widgetIds := getWidgetIdsFromAttachments(comment.Attachments)

				for _, widgetId := range widgetIds {
					if _, exists := widgetMap[widgetId]; !exists {
						widgetMap[widgetId] = true
					}
				}
			}
		}
	case []responses.CommentWithParentResponse:
		for _, comment := range comments {

			tempAttachments := comment.Attachments

			if comment.ParentComment != nil {
				tempAttachments = append(tempAttachments, comment.ParentComment.Attachments...)
			}

			widgetIds := getWidgetIdsFromAttachments(tempAttachments)

			for _, widgetId := range widgetIds {
				if _, exists := widgetMap[widgetId]; !exists {
					widgetMap[widgetId] = true
				}
			}
		}
	case responses.FetchCommentResponse:

		tempAttachments := comments.Attachments

		if comments.ParentComment != nil {
			tempAttachments = append(tempAttachments, comments.ParentComment.Attachments...)
		}

		for _, comment := range comments.Replies {
			tempAttachments = append(tempAttachments, comment.Attachments...)
		}

		widgetIds := getWidgetIdsFromAttachments(tempAttachments)

		for _, widgetId := range widgetIds {
			if _, exists := widgetMap[widgetId]; !exists {
				widgetMap[widgetId] = true
			}
		}
	}

	return widgetMap
}

func typeAssertAndFetchWidgetIdsFromTopics(topics interface{}, widgetMap map[string]bool) map[string]bool {

	switch topics := topics.(type) {
	case responses.TopicResponse:
		if topics.WidgetId != "" {
			if _, exists := widgetMap[topics.WidgetId]; !exists {
				widgetMap[topics.WidgetId] = true
			}
		}
	case responses.TopicResponseWithMeta:
		if topics.WidgetId != "" {
			if _, exists := widgetMap[topics.WidgetId]; !exists {
				widgetMap[topics.WidgetId] = true
			}
		}
	case map[string]responses.TopicResponse:
		for _, topic := range topics {
			if topic.WidgetId != "" {
				if _, exists := widgetMap[topic.WidgetId]; !exists {
					widgetMap[topic.WidgetId] = true
				}
			}
		}
	case []responses.TopicResponse:
		for _, topic := range topics {
			if topic.WidgetId != "" {
				if _, exists := widgetMap[topic.WidgetId]; !exists {
					widgetMap[topic.WidgetId] = true
				}
			}
		}
	case map[string]responses.TopicResponseWithMeta:
		for _, topic := range topics {
			if topic.WidgetId != "" {
				if _, exists := widgetMap[topic.WidgetId]; !exists {
					widgetMap[topic.WidgetId] = true
				}
			}
		}
	case map[string][]responses.TopicResponseWithMeta: // for child_topics
		for _, topic := range topics {
			for _, topic := range topic {
				if topic.WidgetId != "" {
					if _, exists := widgetMap[topic.WidgetId]; !exists {
						widgetMap[topic.WidgetId] = true
					}
				}
			}
		}
	case []responses.TopicResponseWithMeta:
		for _, topic := range topics {
			if topic.WidgetId != "" {
				if _, exists := widgetMap[topic.WidgetId]; !exists {
					widgetMap[topic.WidgetId] = true
				}
			}
		}
	}

	return widgetMap
}

func typeAssertAndFetchWidgetIdsFromPosts(posts interface{}, widgetMap map[primitive.ObjectID]bool) map[primitive.ObjectID]bool {
	switch posts := posts.(type) {
	case []responses.PostResponse:
		for _, post := range posts {
			widgetIds := getWidgetIdsFromAttachments(post.Attachments)

			for _, widgetId := range widgetIds {
				if _, exists := widgetMap[widgetId]; !exists {
					widgetMap[widgetId] = true
				}
			}
		}
	case map[string]responses.PostResponse:
		for _, post := range posts {
			widgetIds := getWidgetIdsFromAttachments(post.Attachments)

			for _, widgetId := range widgetIds {
				if _, exists := widgetMap[widgetId]; !exists {
					widgetMap[widgetId] = true
				}
			}
		}
	case responses.PostResponse:
		widgetIds := getWidgetIdsFromAttachments(posts.Attachments)

		for _, widgetId := range widgetIds {
			if _, exists := widgetMap[widgetId]; !exists {
				widgetMap[widgetId] = true
			}
		}

	case responses.PostWithRepliesResponse:

		tempAttachments := posts.Attachments

		for _, reply := range posts.Replies {
			tempAttachments = append(tempAttachments, reply.Attachments...)
		}

		widgetIds := getWidgetIdsFromAttachments(tempAttachments)

		for _, widgetId := range widgetIds {
			if _, exists := widgetMap[widgetId]; !exists {
				widgetMap[widgetId] = true
			}
		}
	}

	return widgetMap
}

// Internal Method to parse widget_ids from posts
func getWidgetIdsFromPosts(response interface{}) []primitive.ObjectID {
	uniqueWidgetIds := []primitive.ObjectID{}
	tempWidgetIds := map[primitive.ObjectID]bool{}

	// Widgets for single post
	if post, ok := response.(gin.H)["post"]; ok {
		tempWidgetIds = typeAssertAndFetchWidgetIdsFromPosts(post, tempWidgetIds)
	}

	// Widgets for multiple posts
	if posts, ok := response.(gin.H)["posts"]; ok {
		tempWidgetIds = typeAssertAndFetchWidgetIdsFromPosts(posts, tempWidgetIds)
	}

	// Widgets for reposted post
	if repostedPosts, ok := response.(gin.H)["reposted_posts"]; ok {
		tempWidgetIds = typeAssertAndFetchWidgetIdsFromPosts(repostedPosts, tempWidgetIds)
	}

	for key := range tempWidgetIds {
		uniqueWidgetIds = append(uniqueWidgetIds, key)
	}

	return uniqueWidgetIds
}

// Internal Method to get topics Data from Posts response
func getTopicDataFromPosts(topicHelper interfaces.TopicHelper, response interface{}, communityId int) map[string]responses.TopicResponse {
	topicIds := getTopicIdsFromPosts(response)

	topicsData, _ := fetchAndParseTopicsForResponse(topicHelper, topicIds, communityId)

	return topicsData
}

// Internal Method to get widget Data from Posts response
func getWidgetDataFromFeedResponse(handlers *FeedHandlers, response interface{}, communityId int, userIsCM bool, userId string,
) map[string]requests.WidgetResponse {

	// get widget ids from posts
	widgetIds := getWidgetIdsFromPosts(response)

	// get widget ids from topics
	topicWidgetIds := getWidgetIdsFromTopics(response)
	widgetIds = append(widgetIds, topicWidgetIds...)

	// get widget ids from comments
	commentWidgetIds := getWidgetIdsFromComments(response)
	widgetIds = append(widgetIds, commentWidgetIds...)

	// fetch widget data from widget ids
	widgetsData, _ := parseWidgetsResponse(handlers, widgetIds, communityId, userIsCM, userId)

	return widgetsData
}

func getOriginalPostForReposts(handlers *FeedHandlers, response interface{}, communityId int, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool,
) map[string]responses.PostResponse {

	postIds := getPostIdsFromReposts(response, apiRevampV1Check)

	postsData, _ := fetchMultiplePostsData(handlers, postIds, communityId, userId, isCm, versionCode, platformCode, apiRevampV1Check)

	return postsData
}

func getPostIdsFromReposts(response interface{}, apiRevampV1Check bool) []string {
	uniquePostIds := []string{}
	tempPostIds := map[string]bool{}

	// extract from single post {}
	if post, ok := response.(gin.H)["post"]; ok {

		switch postData := post.(type) {
		case responses.PostWithRepliesResponse:
			if postData.IsRepost {
				tempPostIds[postData.Attachments[0].AttachmentMeta.EntityID] = true
			}
		}
	}

	// extract from multiple posts []
	if posts, ok := response.(gin.H)["posts"]; ok {
		switch posts := posts.(type) {
		case []responses.PostResponse:
			for _, post := range posts {
				if post.IsRepost {
					if apiRevampV1Check {
						tempPostIds[string(post.Attachments[0].MetaData.EntityID)] = true
					} else {
						tempPostIds[post.Attachments[0].AttachmentMeta.EntityID] = true
					}
				}
			}
		case map[string]responses.PostResponse:
			for _, post := range posts {
				if post.IsRepost {
					if apiRevampV1Check {
						tempPostIds[string(post.Attachments[0].MetaData.EntityID)] = true
					} else {
						tempPostIds[post.Attachments[0].AttachmentMeta.EntityID] = true
					}
				}
			}
		}
	}

	for key := range tempPostIds {
		uniquePostIds = append(uniquePostIds, key)
	}

	return uniquePostIds
}

// Internal method of adding topics, reposted_posts, widgets data in response
func addMetadataInResponse(handlers *FeedHandlers, response gin.H, communityId int, memberId string, platformCode string,
	versionCode string, userIsCM bool, apiRevampV1Check bool, addTopicsData bool, addRespostedPostsData bool,
	addWidgetsData bool) gin.H {

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, response, communityId, memberId, userIsCM,
		versionCode, platformCode, apiRevampV1Check)
	response["widgets"] = getWidgetDataFromPostsAndTopics(handlers, response, communityId, userIsCM, memberId)

	return response
}

func getOriginalPostIDFromRepostRequest(createPostRequest requests.CreatePostRequest) string {
	for _, attachement := range createPostRequest.Attachments {
		if attachement.AttachmentType == enums.PostWidget {
			return attachement.AttachmentMeta.EntityID
		}
	}
	return ""
}

// updateOriginalPostWidgetForRepost | updates original post's repost widget data for a new repost
func updateOriginalPostWidgetForRepost(handlers *FeedHandlers, originalPostID string, repostID interface{}, repostCreatorUserID string) {

	postFilterData := gin.H{
		"_id": originalPostID,
	}
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
	if err != nil {
		return
	}

	if len(postResults) <= 0 {
		return
	}
	originalPost := postResults[0]

	originalPostRepostWidgetData := getRepostWidgetDataFromPost(originalPost)
	if originalPostRepostWidgetData.AttachmentType == enums.RepostWidget {
		//get respost widget id, update respost widget data
		repostWidgetID := originalPostRepostWidgetData.AttachmentMeta.EntityID

		widgetFilter := gin.H{
			"_id": repostWidgetID,
		}
		repostWidgets, err := handlers.widgetHelper.FindWidgetHelper(widgetFilter, gin.H{})
		if err != nil {
			return
		}

		if len(repostWidgets) <= 0 {
			return
		}

		repostWidgetData := repostWidgets[0]
		repostWidgetMetadata := repostWidgetData.MetaData
		repostWidgetMetadataReposts := repostWidgetMetadata["reposts"]
		repostWidgetMetadataRepostsMap, ok := repostWidgetMetadataReposts.(map[string]interface{})
		if !ok {
			return
		}
		repostWidgetMetadataRepostsMap[repostCreatorUserID] = gin.H{
			"repost_id": repostID.(primitive.ObjectID),
		}

		repostWidgetMetadataRepostCount := repostWidgetMetadata["repost_count"].(int32)
		repostWidgetMetadataRepostCount = repostWidgetMetadataRepostCount + 1

		respostWidgetMetaData := gin.H{
			"reposts":      repostWidgetMetadataRepostsMap,
			"repost_count": repostWidgetMetadataRepostCount,
		}

		widgetUpdateData := gin.H{
			"$set": gin.H{
				"metadata": respostWidgetMetaData,
			},
		}

		// update widget data
		handlers.widgetHelper.UpdateWidgetByIdHelper(repostWidgetID, widgetUpdateData)

		return
	}

	// if repost widget does not exists for the post, create repost widget
	respostWidgetMetaData := gin.H{
		"reposts": gin.H{
			repostCreatorUserID: gin.H{
				"repost_id": repostID.(primitive.ObjectID),
			},
		},
		"repost_count": 1,
	}

	repostWidgetID, err := handlers.widgetHelper.CreateWidgetHelper(true, originalPostID, constants.PostEntityType, respostWidgetMetaData, gin.H{}, originalPost.CommunityId)
	if err != nil {
		return
	}

	repostAttachmentMeta := &entities.AttachmentMeta{
		OgTags:   &entities.OGTags{},
		EntityID: repostWidgetID.(primitive.ObjectID),
	}

	originalPostAttachments := originalPost.Attachments
	repostWidgetAttachmentData := entities.Attachment{
		AttachmentType: enums.RepostType.ToInt(),
		AttachmentMeta: repostAttachmentMeta}

	originalPostAttachments = append(originalPostAttachments, repostWidgetAttachmentData)

	originalPostIDPrimitiveObject, _ := primitive.ObjectIDFromHex(originalPostID)
	postUpdateData := gin.H{
		"$set": gin.H{
			"attachments": originalPostAttachments,
		},
	}

	// save respost widget in original post attachments
	handlers.postHelper.UpdatePostByIdHelper(originalPostIDPrimitiveObject, postUpdateData)
}

// Internal Method to parse post for response
func parsePostResponse(handlers *FeedHandlers, post entities.Post, userId string, isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool, memberRole string,
) responses.PostResponse {

	likes_count, _ := fetchEntityLikesCount(handlers.likeHelper, post.ID.Hex(), constants.PostEntityType)
	replies_count, _ := fetchPostCommentsCount(handlers.commentHelper, post.ID.Hex())

	var response responses.PostResponse

	response.ID = post.ID
	response.TempID = post.TempId
	response.Text = post.Text
	response.Topics = post.TopicIds
	response.Heading = post.Heading
	response.CommunityId = post.CommunityId
	response.ChatroomId = post.ChatroomId
	response.IsPinned = post.IsPinned
	response.UserId = post.UserId
	response.UUID = post.UserId
	response.Attachments = ParseAttachmentsforResponse(post.Attachments, apiRevampV1Check)
	response.LikesCount = int(likes_count)
	response.CommentsCount = int(replies_count)
	response.RepostCount = getPostRepostCount(handlers.widgetHelper, post)
	response.IsDeleted = post.IsDeleted
	response.IsEdited = post.IsEdited
	response.IsRepost = post.IsRepost
	response.IsRepostedByUser = getIsRepostedByUser(handlers.widgetHelper, userId, post)
	response.IsLiked = fetchUserLikedStatusByEntity(handlers.likeHelper, post.ID.Hex(), constants.PostEntityType, userId)
	response.IsSaved = fetchUserSavedStatusByPostId(handlers.saveHelper, post.ID.Hex(), userId)
	response.MenuItems = []responses.MenuResponse{}

	if memberRole != utils.GuestRole {
		response.MenuItems = getEntityMenuItems(constants.PostEntityType, isCm, userId == post.UserId, post.IsPinned,
			versionCode, platformCode, userId, post.CommunityId, handlers.cacheHelper)
	}

	if post.IsDeleted {
		response.DeleteReason = post.DeleteReason
		response.DeletedBy = post.DeletedBy
		response.DeletedByUUID = post.DeletedBy
	}

	response.CreatedAt = int(post.CreatedAt.UnixMilli())
	response.UpdatedAt = int(post.UpdatedAt.UnixMilli())

	if apiRevampV1Check {
		// remove community_id and user_id from post response
		response.CommunityId = 0
		response.UserId = ""

	}

	response.IsPendingPost = false

	return response
}

// Internal Method to parse multiple post for response
func parseMultiplePostResponse(handlers *FeedHandlers, posts []entities.Post, userId string, isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool, memberRole string,
) []responses.PostResponse {

	response := []responses.PostResponse{}

	for _, post := range posts {
		response = append(response, parsePostResponse(handlers, post, userId, isCm, versionCode, platformCode,
			apiRevampV1Check, memberRole))
	}

	return response
}

// Internal Method to fetch post using post_id and community_id
func fetchPost(helper interfaces.PostHelper, postId string, communityId int) (*entities.Post, error) {
	// post filter data
	postFilterData := gin.H{
		"_id":          postId,
		"is_deleted":   false,
		"community_id": communityId,
	}

	// fetch post using helper method
	postResults, err := helper.FindPostHelper(postFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of post_id
	if len(postResults) == 0 {
		return nil, fmt.Errorf("invalid post_id sent")
	}

	return &postResults[0], nil
}

// getPostByID | get post data by id
func getPostByID(helper interfaces.PostHelper, postID string) (*entities.Post, error) {
	filter := gin.H{
		"_id": postID,
	}

	postResults, err := helper.FindPostHelper(filter, gin.H{})
	if err != nil {
		return nil, err
	}

	if len(postResults) == 0 {
		return nil, fmt.Errorf("invalid post_id")
	}

	return &postResults[0], nil
}

// Internal Method to fetch parsed post with replies
func fetchPostWithReplies(handlers *FeedHandlers, postId string, communityId int, filterOptions map[string]interface{},
	memberId string, isCm bool, versionCode string, platformCode string, apiRevampV1Check bool, memberRole string,
) (responses.PostWithRepliesResponse, error) {

	var postWithRepliesResponse responses.PostWithRepliesResponse

	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		return postWithRepliesResponse, err
	}

	commentFilterData := gin.H{
		"level":      constants.CommentBaseLevel,
		"is_deleted": false,
		"post_id":    postId,
	}

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, filterOptions)
	if err != nil {
		return postWithRepliesResponse, err
	}

	postResponse := parsePostResponse(handlers, *postData, memberId, isCm, versionCode, platformCode, apiRevampV1Check, memberRole)
	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, memberId, isCm,
		versionCode, platformCode, apiRevampV1Check, handlers.cacheHelper, memberRole)

	postWithRepliesResponse.PostResponse = postResponse
	postWithRepliesResponse.Replies = repliesResponse

	return postWithRepliesResponse, nil
}

// Internal Method to fetch single post response
func FetchSinglePostResponse(handlers *FeedHandlers, postId string) (*responses.PostResponse, error) {

	postData, err := getPostByID(handlers.postHelper, postId)
	if err != nil {
		return nil, err
	}

	postResponse := parsePostResponse(handlers, *postData, postData.UserId, false, "", "", false, utils.DefaultRole)

	return &postResponse, nil
}

// Internal Method to fetch multiple posts data using post_ids
func fetchMultiplePostsData(handlers *FeedHandlers, postIds []string, communityId int, userId string,
	isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool) (map[string]responses.PostResponse, error) {

	// convert post_ids to object ids
	postObjectIds := helpers.ConvertIdsToObjectIds(postIds)

	// filter options to fetch posts from db
	filterOptions := gin.H{
		"_id": gin.H{
			"$in": postObjectIds,
		},
		"community_id": communityId,
	}

	// fetch posts using helper method
	postsLists, err := handlers.postHelper.FindPostHelper(filterOptions, gin.H{})
	if err != nil {
		return nil, err
	}

	// Make key value pair of post_id -> PostResponse
	postResponse := map[string]responses.PostResponse{}

	// parse post response data for each post
	for _, post := range postsLists {
		postResponse[post.ID.Hex()] = parsePostResponse(handlers, post, userId, isCm, versionCode, platformCode, apiRevampV1Check, utils.DefaultRole)
	}

	return postResponse, nil

}

// Internal function to fetch posts with topic id
func fetchPostsWithTopicID(postHelper interfaces.PostHelper, topicId primitive.ObjectID, communityId int) ([]entities.Post, error) {
	// filter to find posts with the specified topic_id and is_deleted set to false
	filter := bson.M{
		"topic_ids": bson.M{
			"$elemMatch": bson.M{
				"$eq": topicId,
			},
		},
		"is_deleted":   false,
		"community_id": communityId,
	}

	// find posts based on the filter
	postResults, err := postHelper.FindPostHelper(filter, gin.H{})
	if err != nil {
		return nil, err
	}

	return postResults, nil
}

func createActivitiesAndSendNotificationAfterPostCreation(handlers *FeedHandlers, userId string, communityId int,
	headers map[string]string, postRequest requests.CreatePostRequest, postData *entities.Post) error {

	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]

	// create activity for repost
	if postRequest.IsRepost {
		originalPostID := getOriginalPostIDFromRepostRequest(postRequest)

		// create activity for repost
		postFilterData := gin.H{
			"_id": originalPostID,
		}
		postResults, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
		if err != nil {
			return fmt.Errorf("original post not found for repost")
		}

		originalPost := postResults[0]
		OriginalPostUserID := originalPost.UserId
		ctaData := gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     originalPostID,
		}

		OriginalPostIDObject, _ := primitive.ObjectIDFromHex(originalPostID)

		activityID, err := handlers.CreateActivity(communityId, []string{userId}, OriginalPostUserID, constants.Post,
			OriginalPostIDObject, OriginalPostUserID, constants.RepostOnPost, ctaData, false, false, primitive.NilObjectID)
		if err != nil {
			// utils.GeneralAPIInternalError(c, err.Error())
			return err
		}

		if activityID != nil {
			err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), platformCode, versionCode)
			if err != nil {
				logging.Error("Failed to enqueue send notification : ", err)
			}
		}
	}

	// create activity for tagged members
	if len(postRequest.UUIDs) > 0 {

		// cta data for activity
		ctaData := gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     postData.ID.Hex(),
		}

		for _, member := range postRequest.UUIDs {
			// create tag activity
			activityID, err := handlers.CreateActivity(communityId, []string{userId}, member, constants.Post,
				postData.ID, userId, constants.TaggedInPost, ctaData, false, false, primitive.NilObjectID)
			if err != nil {
				return err
			}

			if activityID != nil {
				err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
				if err != nil {
					logging.Error("Failed to enqueue send notification : ", err)
				}
			}

		}
	}

	return nil
}

func createNormalPostAfterValidation(handlers *FeedHandlers, userId string, communityId int,
	postRequest *requests.CreatePostRequest, headers map[string]string) (*entities.Post, error) {

	// create post using the helper method
	postId, err := handlers.postHelper.CreatePostHelper(postRequest.Text, postRequest.Heading,
		communityId, userId, postRequest.Attachments, postRequest.ChatroomID,
		postRequest.TempID, postRequest.ParsedTopicIds, postRequest.OriginalAuthor, postRequest.Visibility,
		postRequest.IsRepost, postRequest.CreatedAt)
	if err != nil {
		return nil, err
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, postRequest.PostType, postRequest.Attachments,
		postId.(primitive.ObjectID).Hex(), communityId, userId)
	if err != nil {
		return nil, err
	}

	// update post data using helper method
	err = handlers.postHelper.EditPostHelper(postId.(primitive.ObjectID), postRequest.Text,
		postRequest.Heading, updatedAttachments, postRequest.ParsedTopicIds, postRequest.Visibility, false)
	if err != nil {
		return nil, err
	}

	// update post in connection buffer lists
	userConnectionData, _ := getUserConnectionDataFromCache(handlers, userId, communityId)
	if len(userConnectionData) == 0 {
		updateConnectionList(handlers, userId, communityId, "", false)
	}

	userConnectionData, _ = getUserConnectionDataFromCache(handlers, userId, communityId)
	for connectionData := range userConnectionData {
		updateConnectionFeedBuffer(handlers, connectionData, communityId, postId.(primitive.ObjectID).Hex(), true)
	}

	// fetch post data using new post_id
	postData, err := fetchPost(handlers.postHelper, postId.(primitive.ObjectID).Hex(), communityId)
	if err != nil {
		return nil, err
	}

	// insert post data in elastic search
	err = handlers.esHelper.IndexDocument(ParsePostIndexData(postData), postData.ID.Hex(),
		constants.PostIndexName)
	if err != nil {
		logging.Error(fmt.Sprint("Error in inserting post data in elastic search: ", err.Error()))
	}

	// Update posts count in topics index
	if len(postData.TopicIds) > 0 {

		stringTopicIds := helpers.ParseObjectIdsToString(postData.TopicIds)
		updatePostCountInTopicQuery := UpdatePostCountInTopicsQuery(stringTopicIds, true)

		err = handlers.esHelper.UpdateByQuery(updatePostCountInTopicQuery, constants.TopicIndexName)
		if err != nil {
			logging.Error(err.Error())
		}
	}

	// update original post widget for repost
	if postRequest.IsRepost {
		originalPostID := getOriginalPostIDFromRepostRequest(*postRequest)
		updateOriginalPostWidgetForRepost(handlers, originalPostID, postData.ID, userId)
	}

	// Task after creation of normal post
	tasksAfterPostCreation(handlers, postData, *postRequest, userId, communityId, headers)

	return postData, nil
}

// Internal method to create normal or pending post after validation of request
func createPostAfterValidation(handlers *FeedHandlers, userId string, communityId int,
	postRequest *requests.CreatePostRequest, headers map[string]string) (*entities.Post, error) {

	var postData *entities.Post
	var err error

	// create post based on post type
	if postRequest.PostType == constants.PendingPostEntityType {
		postData, err = createPendingPostAfterValidation(handlers, userId, communityId, postRequest)
	} else {
		postData, err = createNormalPostAfterValidation(handlers, userId, communityId, postRequest, headers)
	}

	return postData, err
}

func validateCreatePostRequest(handlers *FeedHandlers, userId string, communityId int, apiRevampV1Check bool, postRequest *requests.CreatePostRequest) (gin.H, error) {

	postRequest.Text = strings.Trim(postRequest.Text, " ")
	if postRequest.Text == "" && postRequest.Heading == "" && len(postRequest.Attachments) == 0 {
		return nil, fmt.Errorf("can't create post without content")
	}

	// validation of create post attachments
	err := ValidateAndUpdateAttachments(handlers, communityId, enums.EntityTypePost, postRequest.Attachments, apiRevampV1Check, false, postRequest.IsRepost)
	if err != nil {
		return nil, err
	}

	if postRequest.IsRepost {
		originalPostID := getOriginalPostIDFromRepostRequest(*postRequest)
		err := validateUserAndPostForRepost(handlers, userId, originalPostID, communityId)
		if err != nil {
			return nil, err
		}
	}

	// convert topic_ids to object ids
	topicIDs := helpers.ConvertIdsToObjectIds(postRequest.TopicIds)

	// validate if topic_ids are valid
	if len(topicIDs) > 0 {
		topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, false)
		if err != nil {
			return nil, err
		}

		// Validation of Topics
		if len(topics) != len(topicIDs) {
			return nil, fmt.Errorf("invalid topic_ids sent")
		}
	}

	// update parsed topic ids in request struct
	postRequest.ParsedTopicIds = topicIDs

	// check the visibility of the post
	if postRequest.Visibility == "" {
		postRequest.Visibility = enums.PublicVisibility
	}

	if postRequest.Visibility != enums.PrivateVisibility && postRequest.Visibility != enums.PublicVisibility {
		// utils.GeneralAPIValidationError(c, "Invalid visibility sent")
		return nil, fmt.Errorf("invalid visibility sent")
	}

	// If NSFW Filtering is enabled & attachments are present, check for NSFW content
	if len(postRequest.Attachments) > 0 {
		errorMeta, err := validateAndUpdatePostImagesForNSFWContent(handlers.cacheHelper, userId, communityId,
			&postRequest.Attachments, nil)
		if errorMeta != nil {
			return errorMeta, err
		}
	}

	return nil, nil
}

// Exposed Method to create a Post
func (handlers *FeedHandlers) CreatePost(c *gin.Context) {

	// fetch headers
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	memberRole := headers[utils.HeaderMemberRole]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// bind request body to struct
	var createPostRequest requests.CreatePostRequest
	if err := c.ShouldBindJSON(&createPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// if on_behalf_of_uuid is not empty
	if createPostRequest.OnBehalfOfUUID != "" {
		// Validate if user is cm or not
		if !createPostRequest.UserIsCm {
			utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		}

		// update UserId and OriginalAuthorUUID
		createPostRequest.OriginalAuthor = userId
		userId = createPostRequest.OnBehalfOfUUID
	}

	// Validate create post request
	errorMeta, err := validateCreatePostRequest(handlers, userId, communityId, apiRevampV1Check, &createPostRequest)
	if err != nil {

		// if errorMeta is not nil, return custom error with meta else return validation error
		if errorMeta != nil {
			utils.CustomAPIErrorWithMeta(c, http.StatusBadRequest, err.Error(), errorMeta)
		} else {
			utils.GeneralAPIValidationError(c, err.Error())
		}
		return
	}

	// create normal post using internal method
	createPostRequest.PostType = constants.PostEntityType
	postData, err := createPostAfterValidation(handlers, userId, communityId, &createPostRequest, headers)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options for pagination
	filterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch post response data
	fetchPostData, err := fetchPostWithReplies(handlers, postData.ID.Hex(), communityId,
		filterOptions, headers[utils.HeadersMemberId], createPostRequest.UserIsCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check, memberRole)
	if err == nil {
		response["post"] = fetchPostData
		response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
		response["reposted_posts"] = getOriginalPostForReposts(handlers, response, communityId, headers[utils.HeadersMemberId],
			createPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
		response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, createPostRequest.UserIsCm, headers[utils.HeadersMemberId])
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

func tasksAfterPostCreation(handlers *FeedHandlers, postData *entities.Post, postRequest requests.CreatePostRequest, userId string,
	communityId int, headers map[string]string) {

	apiKey := headers[utils.HeadersApiKey]

	// Add the post topics data in PostTopics collection
	if postData.TopicIds != nil {
		// Trigger create post background tasks
		err := handlers.taskDistributor.AsyncCreatePostTasks(postData.ID.Hex())
		if err != nil {
			logging.Error("Error enqueuing create post background task", err)
		}
	}

	// if custom creation timestamp is not used, create activities and send notifications
	if !(postRequest.CreatedAt > 0 && float64(postRequest.CreatedAt) <= float64(time.Now().UnixMilli())) {
		err := createActivitiesAndSendNotificationAfterPostCreation(handlers, userId, communityId, headers, postRequest, postData)
		if err != nil {
			logging.Error(fmt.Sprint("Error while creating activities and sending notifications after post creation: ", err.Error()))
		}

		// trigger post creation webhook
		err = handlers.taskDistributor.TriggerPostCreationWebhook(postData.ID.Hex(), apiKey)
		if err != nil {
			logging.Error("Error while triggering post creation webhook: ", err.Error())
		}

		// trigger post tagged webhook
		if len(postRequest.UUIDs) > 0 {
			err = handlers.taskDistributor.TriggerPostTaggedWebhook(postData.ID.Hex(), postRequest.UUIDs, apiKey)
			if err != nil {
				logging.Error("Error while triggering post tagged webhook: ", err.Error())
			}
		}
	}
}

// Exposed Method to fetch multiple posts from post_ids
func (handlers *FeedHandlers) FetchPosts(c *gin.Context) {

	// fetch headers
	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	var fetchPostQueryRequest requests.FetchPostsQueryRequest

	err := c.BindQuery(&fetchPostQueryRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If user is not cm, return error
	if !fetchPostQueryRequest.UserIsCm {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Unmarshal post and pending post ids
	postIds, pendingPostIds := []string{}, []string{}

	if fetchPostQueryRequest.PostIds != "" {
		err := json.Unmarshal([]byte(fetchPostQueryRequest.PostIds), &postIds)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}
	}

	if fetchPostQueryRequest.PendingPostIds != "" {
		err := json.Unmarshal([]byte(fetchPostQueryRequest.PendingPostIds), &pendingPostIds)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}
	}

	postsResponse := map[string]responses.PostResponse{}

	if len(postIds) > 0 {
		// fetch multiple posts data using internal method
		postsResponse, err = fetchMultiplePostsData(handlers, postIds, communityId, headers[utils.HeadersMemberId],
			true, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

	}

	// If pending_post_ids are present, fetch posts data from pending posts
	if len(pendingPostIds) > 0 {

		// Fetch posts data from pending posts using internal method
		pendingPostData, err := fetchMultiplePendingPostsData(handlers, pendingPostIds, communityId, headers[utils.HeadersMemberId],
			true, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// Add parsed posts data to response
		for key, value := range pendingPostData {
			postsResponse[key] = value
		}

	}

	// reponse data
	response := gin.H{
		"posts":   postsResponse,
		"success": true,
	}

	postsList := []responses.PostResponse{}
	for _, value := range postsResponse {
		postsList = append(postsList, value)
	}

	parsedResponse := gin.H{
		"posts": postsList,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, parsedResponse, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, response, communityId, headers[utils.HeadersMemberId],
		fetchPostQueryRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, parsedResponse, communityId, fetchPostQueryRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to fetch a Post using post_id
func (handlers *FeedHandlers) FetchPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	memberRole := headers[utils.HeaderMemberRole]

	if paramIsCm == "true" {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch post response data
	fetchPostData, err := fetchPostWithReplies(handlers, postId, communityId, commentFilterOptions,
		headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check, memberRole)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}
	response["post"] = fetchPostData
	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, response, communityId, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, isCm, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to edit a Post
func (handlers *FeedHandlers) EditPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	memberRole := headers[utils.HeaderMemberRole]

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editPostRequest requests.EditPostRequest
	if err := c.ShouldBindJSON(&editPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Check if user is cm or post creator
	if !editPostRequest.UserIsCm && postData.UserId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// validation of attachment objects
	err = ValidateAndUpdateAttachments(handlers, communityId, enums.EntityTypePost, editPostRequest.Attachments, apiRevampV1Check, true, postData.IsRepost)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validates a respost's post attachement in edit request
	if postData.IsRepost && !validateRepostPostAttachment(postData, editPostRequest) {
		utils.GeneralAPIValidationError(c, "cannot update repost's post attachment")
		return
	}

	// If NSFW Filtering is enabled & attachments are present, check for NSFW content
	if len(editPostRequest.Attachments) > 0 {
		errorMeta, err := validateAndUpdatePostImagesForNSFWContent(handlers.cacheHelper, headers[utils.HeadersMemberId], communityId,
			&editPostRequest.Attachments, &postData.Attachments)
		if errorMeta != nil {
			utils.CustomAPIErrorWithMeta(c, http.StatusBadRequest, err.Error(), errorMeta)
			return
		}
	}

	// strip text and check if it is empty
	editPostRequest.Text = strings.TrimSpace(editPostRequest.Text)

	if editPostRequest.Text == "" && editPostRequest.Heading == "" && len(editPostRequest.Attachments) == 0 {
		utils.GeneralAPIValidationError(c, "Can't Edit post without content")
		return
	}

	topicIDs := postData.TopicIds
	existingTopicIds := postData.TopicIds

	// fetch all the topics sent in the edit post body
	if editPostRequest.TopicIds != nil {
		// convert topic_ids to object ids
		topicIDs = helpers.ConvertIdsToObjectIds(editPostRequest.TopicIds)

		topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, false)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		// Validation of Topics
		if len(topics) != len(topicIDs) {
			utils.GeneralAPIValidationError(c, "Invalid topic_ids sent")
			return
		}
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, constants.PostEntityType, editPostRequest.Attachments,
		postId, communityId, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// check the visibility of the post
	if editPostRequest.Visibility == "" {
		editPostRequest.Visibility = enums.PublicVisibility
	}

	if editPostRequest.Visibility != enums.PrivateVisibility && editPostRequest.Visibility != enums.PublicVisibility {
		utils.GeneralAPIValidationError(c, "Invalid visibility sent")
		return
	}

	// update post data using helper method
	err = handlers.postHelper.EditPostHelper(postData.ID, editPostRequest.Text, editPostRequest.Heading, updatedAttachments,
		topicIDs, editPostRequest.Visibility, true)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post response data
	fetchPostData, err := fetchPostWithReplies(handlers, postId, communityId, commentFilterOptions, headers[utils.HeadersMemberId],
		editPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check, memberRole)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	postData, err = fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Update the post topics data
	if postData.TopicIds != nil {
		// Trigger edit post background tasks
		err = handlers.taskDistributor.AsyncEditPostTasks(postData.ID.Hex())
		if err != nil {
			logging.Error("Error enqueing edit post background task", err)
		}
	}

	// update post data in elastic search
	err = handlers.esHelper.IndexDocument(ParsePostIndexData(postData), postData.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	if editPostRequest.TopicIds != nil {
		updatePostCountInTopics(handlers, editPostRequest.TopicIds, existingTopicIds)
	}

	response := gin.H{
		"success": true,
		"post":    fetchPostData,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, response, communityId, headers[utils.HeadersMemberId], editPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, editPostRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)

}

// updates the count of post in topics
func updatePostCountInTopics(handlers *FeedHandlers, editRequestTopicIds []string, existingTopicIds []primitive.ObjectID) {
	updatedTopicIds := helpers.ConvertIdsToObjectIds(editRequestTopicIds)

	// topics added in the post
	addedTopicIds := utils.GetDifferenceBetweenArray(updatedTopicIds, existingTopicIds)

	// topics removed from the post
	removedTopicIds := utils.GetDifferenceBetweenArray(existingTopicIds, updatedTopicIds)

	// update the count of posts in added topics
	if len(addedTopicIds) > 0 {
		stringTopicIds := helpers.ParseObjectIdsToString(addedTopicIds)
		err := handlers.esHelper.UpdateByQuery(UpdatePostCountInTopicsQuery(stringTopicIds, true), constants.TopicIndexName)
		if err != nil {
			logging.Error(err.Error())
		}
	}

	// update the count of posts in removed topics
	if len(removedTopicIds) > 0 {
		stringTopicIds := helpers.ParseObjectIdsToString(removedTopicIds)
		err := handlers.esHelper.UpdateByQuery(UpdatePostCountInTopicsQuery(stringTopicIds, false), constants.TopicIndexName)
		if err != nil {
			logging.Error(err.Error())
		}
	}
}

// Exposed Method to delete a Post
func (handlers *FeedHandlers) DeletePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deletePostRequest requests.DeletePostRequest
	if err := c.ShouldBindJSON(&deletePostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post using helper method
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deletePostRequest.UserIsCm && headers[utils.HeadersMemberId] != postData.UserId {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// update data
	updateData := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": deletePostRequest.DeleteReason,
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(postData.ID, updateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// soft delete post comments
	commentUpdateData := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": constants.CommentDeleteReasonForPostDelete,
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	commentFilter := gin.H{
		"post_id":    postData.ID,
		"is_deleted": false,
	}

	err = handlers.commentHelper.UpdateManyCommentsHelper(commentFilter, commentUpdateData)
	if err != nil {
		logging.Error("Error updating post comments: ", err.Error())
	}

	// if repost, remove repost data from original post's repost widget
	if postData.IsRepost {
		deleteOriginalPostRepostWidgetData(handlers, postData)
		deleteUserPostRepostActivity(handlers, postData, headers)
	}

	// delete post data in elastic search
	err = handlers.esHelper.DeleteDocument(postData.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// remove activity for post comments
	handlers.removePostCommentActivityData(postData.ID)

	// remove activity for the post
	deleteActivityFilter := gin.H{
		"entity_type": constants.Post,
		"entity_id":   postData.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// delete and fill cache data
	handlers.activityHelper.WarmupUserActivityFeedCache(postData.CommunityId, postData.UserId)
	handlers.activityHelper.WarmupUniversalFeedCache(postData.CommunityId)

	// if deleted by CM, create delete activity
	if deletePostRequest.UserIsCm && headers[utils.HeadersMemberId] != postData.UserId {
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
			postData.UserId, constants.Post, postData.ID, postData.UserId, constants.CMDeletedPost, gin.H{},
			false, false, primitive.NilObjectID)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
			if err != nil {
				logging.Error("Failed to enqueue send notification : ", err)
			}
		}
	}

	// update the count of posts in topics
	if len(postData.TopicIds) > 0 {
		stringTopicIds := helpers.ParseObjectIdsToString(postData.TopicIds)
		err = handlers.esHelper.UpdateByQuery(UpdatePostCountInTopicsQuery(stringTopicIds, false), constants.TopicIndexName)
		if err != nil {
			logging.Error(err.Error())
		}

		// Trigger delete post background tasks
		err = handlers.taskDistributor.AsyncDeletePostTasks(postData.ID.Hex())
		if err != nil {
			logging.Error("Error enqueing delete post background task", err)
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func deleteOriginalPostRepostWidgetData(handlers *FeedHandlers, postData *entities.Post) {
	PostAttachmentData := getPostAttachmentDataFromPost(*postData)
	if PostAttachmentData.AttachmentType != enums.PostWidget {
		return
	}
	OriginalPostID := PostAttachmentData.AttachmentMeta.EntityID

	postFilter := gin.H{
		"_id": OriginalPostID,
	}

	postDatas, err := handlers.postHelper.FindPostHelper(postFilter, gin.H{})
	if err != nil || len(postDatas) <= 0 {
		return
	}

	originalPostData := postDatas[0]
	RepostWidgetData := getRepostWidgetDataFromPost(originalPostData)
	if RepostWidgetData.AttachmentType != enums.RepostWidget {
		return
	}
	//get repost widget id, update repost widget data
	repostWidgetID := RepostWidgetData.AttachmentMeta.EntityID

	widgetFilter := gin.H{
		"_id": repostWidgetID,
	}
	repostWidgets, err := handlers.widgetHelper.FindWidgetHelper(widgetFilter, gin.H{})
	if err != nil || len(repostWidgets) <= 0 {
		return
	}

	repostWidgetData := repostWidgets[0]
	repostWidgetMetadata := repostWidgetData.MetaData
	repostWidgetMetadataReposts := repostWidgetMetadata["reposts"]
	repostWidgetMetadataRepostsMap, ok := repostWidgetMetadataReposts.(map[string]interface{})
	if !ok {
		return
	}

	delete(repostWidgetMetadataRepostsMap, postData.UserId)

	repostWidgetMetadataRepostCount := repostWidgetMetadata["repost_count"].(int32)
	repostCount := repostWidgetMetadataRepostCount - 1
	if repostCount < 0 {
		repostCount = 0
	}
	repostWidgetMetadataRepostCount = repostCount

	respostWidgetMetaData := gin.H{
		"reposts":      repostWidgetMetadataRepostsMap,
		"repost_count": repostWidgetMetadataRepostCount,
	}

	widgetUpdateData := gin.H{
		"$set": gin.H{
			"metadata": respostWidgetMetaData,
		},
	}

	// update widget data
	handlers.widgetHelper.UpdateWidgetByIdHelper(repostWidgetID, widgetUpdateData)

	return
}

func deleteUserPostRepostActivity(handlers *FeedHandlers, repostPostData *entities.Post, headers map[string]string) error {

	OriginalPostID := repostPostData.Attachments[0].AttachmentMeta.EntityID

	activityFilterData := gin.H{
		"community_id": repostPostData.CommunityId,
		"entity_type":  constants.Post,
		"entity_id":    OriginalPostID,
		"action":       constants.RepostOnPost,
	}

	activity, err := handlers.activityHelper.FindActivityHelper(activityFilterData, gin.H{})
	if err != nil {
		return err
	}

	if activity == nil {
		return errors.New("activity not found")
	}

	// remove uuid from repost action list
	actionBy := utils.RemoveAllOccurenceStringList(activity[0].ActionBy, headers[utils.HeadersMemberId])

	// remove action by metadata
	actionByMetadata := activity[0].ActionByMetadata
	delete(actionByMetadata, headers[utils.HeadersMemberId])

	// activity update data
	activityUpdateData := gin.H{
		"$set": gin.H{
			"action_by":          actionBy,
			"action_by_metadata": actionByMetadata,
		},
	}

	// update activity data, exisiting activity timestamp remains same to maintain order
	err = handlers.activityHelper.UpdateActivityByIDHelper(activity[0].ID, activityUpdateData, true, true)
	if err != nil {
		return err
	}

	// if action by is [], no user repost on post, mark activity as deleted
	if len(actionBy) == 0 {
		handlers.activityHelper.DeleteActivityHelper(activityFilterData)
	}

	return nil
}

func (handlers *FeedHandlers) removePostCommentActivityData(postID primitive.ObjectID) {
	commentsFilter := gin.H{
		"post_id": postID,
	}

	// fetch comments using helper method
	comments, err := handlers.commentHelper.FindCommentHelper(commentsFilter, nil)
	if err != nil {
		return
	}

	postCommentIds := []primitive.ObjectID{}

	for _, comment := range comments {
		postCommentIds = append(postCommentIds, comment.ID)
	}

	// remove activity for the comment
	deleteActivityFilter := gin.H{
		"entity_type": constants.Comment,
		"entity_id": gin.H{
			"$in": postCommentIds,
		},
	}

	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)
}

// Exposed Method to pin a Post
func (handlers *FeedHandlers) PinPost(c *gin.Context) {

	headers := utils.GetHeaders(c)

	// fetch url params
	postId := c.Param("post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post using helper method
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update data
	updateData := gin.H{
		"$set": gin.H{
			"is_pinned": !postData.IsPinned,
		},
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(postData.ID, updateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch updated post data using post_id
	postData, err = fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	handlers.updatePinnedPostCache(postData)

	// trigger post pinned webhook
	if postData.IsPinned {
		err = handlers.taskDistributor.TriggerPostPinnedWebhook(postData.ID.Hex(), headers[utils.HeadersMemberId], headers[utils.HeadersApiKey])
		if err != nil {
			logging.Error("Error while triggering post pinned webhook: ", err.Error())
		}
	}

	// update post data in elastic search
	err = handlers.esHelper.IndexDocument(ParsePostIndexData(postData), postData.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// Exposed Method to fetch all the Posts created by a User
func (handlers *FeedHandlers) FetchUserCreatedPosts(c *gin.Context) {
	// fetch url params and headers
	headers := utils.GetHeaders(c)
	userId := c.Param("user_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]

	if paramIsCm == "true" {
		isCm = true
	}

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

	// filter options
	postFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post using helper method
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, postFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	createdPostResponse := parseMultiplePostResponse(handlers, postResults, userId, isCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check, utils.DefaultRole)

	response := parseFetchMultiplePostResponse(createdPostResponse, postsCount)

	// response data
	finalResponse := gin.H{
		"posts":   response.Posts,
		"success": response.Success,
	}

	if response.TotalCount > 0 {
		finalResponse["total_count"] = response.TotalCount
	}

	finalResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalResponse, communityId)
	finalResponse["reposted_posts"] = getOriginalPostForReposts(handlers, finalResponse, communityId, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	finalResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalResponse, communityId, isCm, headers[utils.HeadersMemberId])

	// Get community configurations
	universalFeedConfig := externalHelpers.GetUniversalFeedConfigurationsData(handlers.cacheHelper, userId, communityId)

	var commentSortOrderVal int
	filtered_comments := map[string]responses.CommentWithParentResponse{}

	if universalFeedConfig.CommentSortOrder == enums.DescendingSortOrder {
		commentSortOrderVal = -1
	} else {
		commentSortOrderVal = 1
	}

	if universalFeedConfig.CommentSortOn == enums.UniversalFeedTopLikedComments {
		var updatedPostsWithComments []responses.PostResponse
		updatedPostsWithComments, filtered_comments, err = getTopCommentsAgainstPostsSortOnLikes(handlers,
			response.Posts, userId, isCm, communityId, commentSortOrderVal, universalFeedConfig.CommentCount,
			versionCode, platformCode, apiRevampV1Check, utils.DefaultRole)

		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		if len(updatedPostsWithComments) > 0 {
			finalResponse["posts"] = updatedPostsWithComments
		}

	}

	finalResponse["filtered_comments"] = filtered_comments

	// return final response
	c.JSON(http.StatusOK, finalResponse)
}

func processPostSearchData(handlers *FeedHandlers, data map[string]interface{}, userId string,
	isCm bool, versionCode string, platformCode string, apiRevampV1Check bool) []responses.PostResponse {
	postDetails := data["hits"].(map[string]interface{})["hits"].([]interface{})
	var postList []entities.Post

	for _, data := range postDetails {
		postData := data.(map[string]interface{})["_source"].(map[string]interface{})
		postData["_id"] = postData["id"]

		// convert the data to post entity
		var post entities.Post
		b, _ := json.Marshal(postData)
		json.Unmarshal(b, &post)

		postList = append(postList, post)
	}

	postResponse := parseMultiplePostResponse(handlers, postList, userId, isCm, versionCode, platformCode,
		apiRevampV1Check, utils.DefaultRole)

	return postResponse
}

// Exposed Method to search Posts
func (handlers *FeedHandlers) SearchPost(c *gin.Context) {
	// fetch query params and headers
	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	var searchPostRequest requests.SearchPostRequest

	err := c.BindQuery(&searchPostRequest)
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

	// parsing of chatroom ids
	excludedChatroomIds := utils.ParseIntArrayParam(searchPostRequest.ExcludedChatroomIDs)
	parsedExcludedChatroomIds, _ := json.Marshal(excludedChatroomIds)

	// dsl query to search posts
	postQuery := GetPostFilterQuery(page, pageSize, searchPostRequest.SearchType,
		searchPostRequest.Search, fmt.Sprintf("%v", string(parsedExcludedChatroomIds)), communityId)
	response := handlers.esHelper.ExecuteQuery(postQuery, constants.PostIndexName)

	finalResponse := processPostSearchData(handlers, response, headers[utils.HeadersMemberId],
		searchPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check)

	finalParsedResponse := gin.H{
		"success": true,
		"posts":   finalResponse,
	}

	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, communityId)
	finalParsedResponse["reposted_posts"] = getOriginalPostForReposts(handlers, finalParsedResponse, communityId, headers[utils.HeadersMemberId], false, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	finalParsedResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalParsedResponse, communityId, searchPostRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}

// Exposed Method to search user created Posts
func (handlers *FeedHandlers) SearchUserCreatedPost(c *gin.Context) {
	// fetch query params and headers
	userId := c.Param("user_id")
	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	var searchPostRequest requests.SearchPostRequest

	err := c.BindQuery(&searchPostRequest)
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

	if userId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// dsl query to search posts
	postQuery := GetSelfPostFilterQuery(page, pageSize, searchPostRequest.SearchType,
		searchPostRequest.Search, userId, communityId)
	response := handlers.esHelper.ExecuteQuery(postQuery, constants.PostIndexName)

	finalResponse := processPostSearchData(handlers, response, headers[utils.HeadersMemberId],
		searchPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check)

	finalParsedResponse := gin.H{
		"success": true,
		"posts":   finalResponse,
	}

	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, communityId)
	finalParsedResponse["reposted_posts"] = getOriginalPostForReposts(handlers, finalParsedResponse, communityId, headers[utils.HeadersMemberId], false, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	finalParsedResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalParsedResponse, communityId, searchPostRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}

// updatePinnedPostCache | update pinned post data in cache storage
func (handlers *FeedHandlers) updatePinnedPostCache(postData *entities.Post) {
	if postData.IsPinned {
		handlers.addPostToCommunityPinnedPostCache(postData)
	}
	handlers.removePostFromCommunityPinnedPostsCache(postData.CommunityId, postData.ID.Hex())
}

// addPostToCommunityPinnedPostCache | add post to cache storage
func (handlers *FeedHandlers) addPostToCommunityPinnedPostCache(postData *entities.Post) {
	communityPostPinnedKey := fmt.Sprintf("community_{}_pinned_posts", postData.CommunityId)
	postDataBytes, err := json.Marshal(postData)
	if err != nil {
		return
	}
	postDataString := string(postDataBytes)

	cachePostKey := fmt.Sprintf("post_{}", postData.ID.Hex())

	handlers.cacheHelper.LPush(communityPostPinnedKey, postData.ID.Hex(), -1)
	handlers.cacheHelper.Set(cachePostKey, postDataString, 0)
}

// removePostFromCommunityPinnedPostsCache | add post to cache storage
func (handlers *FeedHandlers) removePostFromCommunityPinnedPostsCache(communityID int, postID string) {
	communityPostPinnedKey := fmt.Sprintf("community_{}_pinned_posts", communityID)

	handlers.cacheHelper.LRem(communityPostPinnedKey, 0, postID)
	handlers.cacheHelper.Del(fmt.Sprintf("post_{}", postID))
}

// Create new post topics in PostTopics collection
func CreateOrUpdatePostTopics(handlers *FeedHandlers, postId string, deleteAllExisting bool) error {

	if postId == "" {
		logging.Error("Invalid post ID!")
		return nil
	}

	if deleteAllExisting {
		DeletePostTopics(handlers, postId)
	}

	postObjectId, err := primitive.ObjectIDFromHex(postId)

	if err != nil {
		logging.Error("Invalid post ID!")
		return nil
	}

	postResults, err := handlers.postHelper.FindPostHelper(gin.H{"_id": postObjectId}, gin.H{})

	if err != nil {
		logging.Error("Error in fetching posts")
		return nil
	}

	var originalPost entities.Post

	if len(postResults) > 0 {
		originalPost = postResults[0]
	}

	if len(originalPost.TopicIds) > 0 {
		// Fetch topics
		topicResults, err := handlers.topicHelper.FindTopicHelper(gin.H{"_id": gin.H{"$in": originalPost.TopicIds}}, gin.H{})
		if err != nil {
			return err
		}

		var allTopicIds []primitive.ObjectID

		for _, topicResult := range topicResults {
			allTopicIds = append(allTopicIds, topicResult.AllParentIds...)
			allTopicIds = append(allTopicIds, topicResult.ID)
		}

		if len(allTopicIds) > 0 {
			var postTopicsMap = map[primitive.ObjectID][]primitive.ObjectID{
				postObjectId: allTopicIds,
			}
			return handlers.postTopicsHelper.CreateOrUpdateManyPostTopicsHelper(postTopicsMap, originalPost.CommunityId)
		}
	}

	return nil
}

// Delete post topics in PostTopics collection
func DeletePostTopics(handlers *FeedHandlers, postId string) error {

	if postId == "" {
		logging.Error("Invalid post ID!")
		return nil
	}

	postObjectId, err := primitive.ObjectIDFromHex(postId)

	if err != nil {
		logging.Error("Invalid post ID!")
		return nil
	}

	filter := gin.H{
		"post_id": postObjectId,
	}

	if err := handlers.postTopicsHelper.DeletePostTopicsHelper(filter); err != nil {
		logging.Error("Error in deleting Post Topics")
		return nil
	}

	return nil
}
