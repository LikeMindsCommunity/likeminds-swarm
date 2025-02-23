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

func getPostRepostCount(widgetHelper interfaces.WidgetHelper, post *entities.Post) int {
	var postRepostCount int = 0

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

		return int(repostWidgets[0].MetaData["repost_count"].(int32))
	}

	return postRepostCount
}

// Internal Method to fetch repost count for multiple posts
func GetRepostCountForMultiplePosts(widgetHelper interfaces.WidgetHelper, posts []entities.Post) map[primitive.ObjectID]int {

	defer utils.Timer("GetRepostCountForMultiplePosts")()

	postRepostCount := make(map[primitive.ObjectID]int, len(posts))

	repostWidgetIds := []primitive.ObjectID{}
	for _, post := range posts {
		postRepostWidgetData := getRepostWidgetDataFromPost(&post)
		if postRepostWidgetData.AttachmentType == enums.RepostWidget {
			repostWidgetIds = append(repostWidgetIds, postRepostWidgetData.AttachmentMeta.EntityID)
		}
	}

	if len(repostWidgetIds) == 0 {
		return postRepostCount
	}

	widgetFilter := gin.H{
		"_id": bson.M{
			"$in": repostWidgetIds,
		},
	}

	repostWidgets, err := widgetHelper.FindWidgetHelper(widgetFilter, gin.H{})
	if err != nil {
		logging.Error("Error while fetching repost widgets for posts", err)
		return postRepostCount
	}

	for _, repostWidget := range repostWidgets {
		postId, err := primitive.ObjectIDFromHex(repostWidget.ParentEntityID)
		if err == nil {
			postRepostCount[postId] = int(repostWidget.MetaData["repost_count"].(int32))
		}
	}

	return postRepostCount
}

func getIsRepostedByUser(widgetHelper interfaces.WidgetHelper, userID string, post *entities.Post) bool {

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

func getIsRepostedByUserForMultiplePosts(widgetHelper interfaces.WidgetHelper, userID string, posts []entities.Post,
) map[primitive.ObjectID]bool {

	defer utils.Timer("getIsRepostedByUserForMultiplePosts")()

	isRepostedByUserMap := make(map[primitive.ObjectID]bool, len(posts))

	repostWidgetIds := []primitive.ObjectID{}
	for _, post := range posts {
		postRepostWidgetData := getRepostWidgetDataFromPost(&post)
		if postRepostWidgetData.AttachmentType == enums.RepostWidget {
			repostWidgetIds = append(repostWidgetIds, postRepostWidgetData.AttachmentMeta.EntityID)
		}
	}

	if len(repostWidgetIds) == 0 {
		return isRepostedByUserMap
	}

	widgetFilter := gin.H{
		"_id": bson.M{
			"$in": repostWidgetIds,
		},
	}

	repostWidgets, err := widgetHelper.FindWidgetHelper(widgetFilter, gin.H{})
	if err != nil {
		logging.Error("Error while fetching repost widgets for posts", err)
		return isRepostedByUserMap
	}

	for _, repostWidget := range repostWidgets {
		repostWidgetMetadataRepostsMap, ok := repostWidget.MetaData["reposts"].(map[string]interface{})
		if !ok {
			continue
		}

		postId, err := primitive.ObjectIDFromHex(repostWidget.ParentEntityID)
		if err != nil {
			continue
		}

		if _, ok := repostWidgetMetadataRepostsMap[userID]; ok {
			isRepostedByUserMap[postId] = true
		}
	}

	return isRepostedByUserMap
}

func validateUserAndPostForRepost(handlers *FeedHandlers, userID string, originalPostID string, communityId int) error {
	postFilterData := gin.H{
		"_id": originalPostID,
	}
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
	if (err != nil) || (len(postResults) <= 0) {
		return fmt.Errorf("original post not found for repost")
	}

	if getIsRepostedByUser(handlers.widgetHelper, userID, &postResults[0]) {
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

// Internal Method to parse topics response
func fetchAndParseTopicsForResponse(topicHelper interfaces.TopicHelper, topicIds []primitive.ObjectID, communityId int,
) (map[string]responses.TopicResponse, error) {

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

	widgetsResponse := map[string]requests.WidgetResponse{}

	if len(widgetIds) == 0 {
		return widgetsResponse, nil
	}

	// Fetch widgets using widget Ids
	widgets, err := fetchWidgetsByIDs(handlers.widgetHelper, widgetIds, communityId)
	if err != nil {
		return nil, err
	}

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
			tempTopicIds = getTopicsIdsFromTopicResponse(post.Topics, tempTopicIds)

		case responses.PostResponse:
			tempTopicIds = getTopicsIdsFromTopicResponse(post.Topics, tempTopicIds)
		}
	}

	if posts, ok := response.(gin.H)["posts"]; ok {

		switch posts := posts.(type) {
		case []responses.PostResponse:
			for _, post := range posts {
				tempTopicIds = getTopicsIdsFromTopicResponse(post.Topics, tempTopicIds)
			}
		case map[string]responses.PostResponse:
			for _, post := range posts {
				tempTopicIds = getTopicsIdsFromTopicResponse(post.Topics, tempTopicIds)
			}
		}
	}

	for key := range tempTopicIds {
		uniqueTopicIds = append(uniqueTopicIds, key)
	}

	return uniqueTopicIds
}

func getTopicsIdsFromTopicResponse(topicIds []primitive.ObjectID, topicMap map[primitive.ObjectID]bool) map[primitive.ObjectID]bool {

	for _, topicId := range topicIds {
		if _, exists := topicMap[topicId]; !exists {
			topicMap[topicId] = true
		}
	}

	return topicMap
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

func appendWidgetIdsFromAttachmentsToMap(attachments []responses.AttachmentResponse, widgetMap map[primitive.ObjectID]bool) map[primitive.ObjectID]bool {

	widgetIds := getWidgetIdsFromAttachments(attachments)

	for _, widgetId := range widgetIds {

		if _, exists := widgetMap[widgetId]; !exists {
			widgetMap[widgetId] = true
		}
	}

	return widgetMap

}

func typeAssertAndFetchWidgetIdsFromComments(comments interface{}, widgetMap map[primitive.ObjectID]bool) map[primitive.ObjectID]bool {

	typeOf := fmt.Sprintf("%T", comments)
	print(typeOf)
	switch comments := comments.(type) {
	case responses.CommentResponse:
		widgetMap = appendWidgetIdsFromAttachmentsToMap(comments.Attachments, widgetMap)

	case responses.CommentWithParentResponse:
		tempAttachments := comments.Attachments

		if comments.ParentComment != nil {
			tempAttachments = append(tempAttachments, comments.ParentComment.Attachments...)
		}

		widgetMap = appendWidgetIdsFromAttachmentsToMap(tempAttachments, widgetMap)

	case []responses.CommentResponse:
		for _, comment := range comments {
			widgetMap = appendWidgetIdsFromAttachmentsToMap(comment.Attachments, widgetMap)
		}

	case map[string]responses.CommentResponse:
		for _, comment := range comments {
			widgetMap = appendWidgetIdsFromAttachmentsToMap(comment.Attachments, widgetMap)
		}

	case map[string][]responses.CommentResponse:
		for _, comment := range comments {
			for _, comment := range comment {
				widgetMap = appendWidgetIdsFromAttachmentsToMap(comment.Attachments, widgetMap)
			}
		}
	case []responses.CommentWithParentResponse:
		for _, comment := range comments {

			tempAttachments := comment.Attachments

			if comment.ParentComment != nil {
				tempAttachments = append(tempAttachments, comment.ParentComment.Attachments...)
			}

			widgetMap = appendWidgetIdsFromAttachmentsToMap(tempAttachments, widgetMap)
		}
	case responses.FetchCommentResponse:

		tempAttachments := comments.Attachments

		if comments.ParentComment != nil {
			tempAttachments = append(tempAttachments, comments.ParentComment.Attachments...)
		}

		for _, comment := range comments.Replies {
			tempAttachments = append(tempAttachments, comment.Attachments...)
		}

		widgetMap = appendWidgetIdsFromAttachmentsToMap(tempAttachments, widgetMap)
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
func getTopicDataFromPosts(topicHelper interfaces.TopicHelper, response interface{}, communityId int,
) map[string]responses.TopicResponse {

	defer utils.Timer("getTopicDataFromPosts")()

	topicsMap := map[string]responses.TopicResponse{}

	topicIds := getTopicIdsFromPosts(response)

	if len(topicIds) > 0 {
		topicsData, err := fetchAndParseTopicsForResponse(topicHelper, topicIds, communityId)
		if err != nil {
			logging.Error("Error while fetching topics for posts: ", err)
		}

		topicsMap = topicsData
	}

	return topicsMap
}

// Internal Method to get widget Data from Posts response
func getWidgetDataFromFeedResponse(handlers *FeedHandlers, response interface{}, communityId int, userIsCM bool, userId string,
) map[string]requests.WidgetResponse {

	defer utils.Timer("getWidgetDataFromFeedResponse")()

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

func getOriginalPostForReposts(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, response interface{},
) map[string]responses.PostResponse {

	defer utils.Timer("getOriginalPostForReposts")()

	postsResponseMap := map[string]responses.PostResponse{}

	postIds := getPostIdsFromReposts(response, loggedInUser.ApiRevampCheckV1)
	if len(postIds) > 0 {
		postsResponse, err := fetchPostResponseMapFromPostIds(handlers, loggedInUser, postIds)
		if err != nil {
			logging.Error("Error while fetching original posts for reposts: ", err)
		}

		postsResponseMap = postsResponse
	}

	return postsResponseMap
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
func addMetadataInResponse(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, response gin.H,
) gin.H {

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, loggedInUser.CommunityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, response)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, loggedInUser.CommunityId, loggedInUser.IsCm, loggedInUser.UserId)

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

	originalPostRepostWidgetData := getRepostWidgetDataFromPost(&originalPost)
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

func getPostUserId(post *entities.Post, loggedInUser *LoggedInUserParams) string {
	if post.IsAnonymous && (post.UserId != loggedInUser.UserId && !loggedInUser.IsCm) {
		return constants.AnonymousUserUserId
	}
	return post.UserId
}

// Internal Method to parse post for response
func parsePostResponse(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, post *entities.Post, postSecondaryData *PostSecondaryDataParams,
) responses.PostResponse {

	defer utils.Timer("parsePostResponse")()

	var response responses.PostResponse

	response.ID = post.ID
	response.TempID = post.TempId
	response.Text = post.Text
	response.Topics = post.TopicIds
	response.Heading = post.Heading
	response.CommunityId = post.CommunityId
	response.ChatroomId = post.ChatroomId
	response.IsPinned = post.IsPinned
	response.IsHidden = post.IsHidden
	response.IsDeleted = post.IsDeleted
	response.IsEdited = post.IsEdited
	response.IsRepost = post.IsRepost
	response.CreatedAt = int(post.CreatedAt.UnixMilli())
	response.UpdatedAt = int(post.UpdatedAt.UnixMilli())
	response.IsPendingPost = false
	response.PostShareCount = post.PostShareCount
	response.IsAnonymous = post.IsAnonymous

	// if post is anonymous and user is not cm or creator, then set anonymous-user
	response.UserId = getPostUserId(post, loggedInUser)
	response.UUID = response.UserId

	response.Attachments = ParseAttachmentsforResponse(post.Attachments, loggedInUser.ApiRevampCheckV1)

	response.LikesCount = int(postSecondaryData.LikesCount)
	response.CommentsCount = int(postSecondaryData.RepliesCount)
	response.RepostCount = int(postSecondaryData.RepostCount)

	response.IsRepostedByUser = postSecondaryData.IsRepostedByUser
	response.IsLiked = postSecondaryData.IsLikedByUser
	response.IsSaved = postSecondaryData.IsSavedByUser
	response.ImpressionCount = postSecondaryData.PostImpressions.ImpressionsCount
	response.ReachCount = postSecondaryData.PostImpressions.ReachCount

	response.MenuItems = []responses.MenuResponse{}
	if loggedInUser.MemberRole != utils.GuestRole {
		response.MenuItems = getEntityMenuItems(constants.PostEntityType, loggedInUser.IsCm, loggedInUser.UserId == post.UserId,
			post.IsPinned, post.IsHidden, loggedInUser.VersionCode, loggedInUser.PlatformCode, loggedInUser.UserId,
			post.CommunityId, handlers.cacheHelper, post.UserId) // Taking the entire time of this function
	}

	if post.IsDeleted {
		response.DeleteReason = post.DeleteReason
		response.DeletedBy = post.DeletedBy
		response.DeletedByUUID = post.DeletedBy
	}

	// if apiRevampV1Check is true, then do not parse community_id and user_id
	if loggedInUser.ApiRevampCheckV1 {
		response.CommunityId = 0
		response.UserId = ""
	}

	return response
}

// Internal method to compute post impressions and reach count
func computePostImpressionReachCount(handlers *FeedHandlers, postIds []primitive.ObjectID) (map[primitive.ObjectID]PostImpressionsData, error) {
	postImpressionReachMap := map[primitive.ObjectID]PostImpressionsData{}

	defer utils.Timer("computePostImpressionReachCount")()

	postImpressionsFilter := []map[string]interface{}{
		gin.H{
			"$match": gin.H{
				"entity_type": enums.EntityTypePost,
				"entity_id": gin.H{
					"$in": postIds,
				},
			},
		},
		gin.H{
			"$group": gin.H{
				"_id": "$entity_id",
				"impressions_count": gin.H{
					"$sum": 1,
				},
				"unique_users_list": gin.H{
					"$addToSet": "$user_id",
				},
			},
		},
		gin.H{
			"$project": gin.H{
				"impressions_count": "$impressions_count",
				"reach_count": gin.H{
					"$size": "$unique_users_list",
				},
			},
		},
	}

	postImpressionsData, err := handlers.userEntityTimestampHelper.AggregateUserEntityTimestampHelper(postImpressionsFilter)
	if err != nil {
		logging.Error("Error in fetching post impressions, err: ", err)
		return nil, err
	}

	for _, postImpressionData := range postImpressionsData {
		postImpressionReachMap[postImpressionData["_id"].(primitive.ObjectID)] = PostImpressionsData{
			ImpressionsCount: int(postImpressionData["impressions_count"].(int32)),
			ReachCount:       int(postImpressionData["reach_count"].(int32)),
		}
	}

	return postImpressionReachMap, nil
}

func parseSinglePostResponse(handlers *FeedHandlers, postData *entities.Post, loggedInUser *LoggedInUserParams,
) responses.PostResponse {

	postImpressionsData, _ := computePostImpressionReachCount(handlers, []primitive.ObjectID{postData.ID})

	postSecondaryData := PostSecondaryDataParams{
		LikesCount:       fetchEntityLikesCount(handlers.likeHelper, postData.ID.Hex(), constants.PostEntityType),
		RepliesCount:     fetchPostCommentsCount(handlers.commentHelper, postData.ID.Hex()),
		RepostCount:      getPostRepostCount(handlers.widgetHelper, postData),
		IsRepostedByUser: getIsRepostedByUser(handlers.widgetHelper, loggedInUser.UserId, postData),
		IsLikedByUser:    fetchUserLikedStatusByEntity(handlers.likeHelper, postData.ID.Hex(), constants.PostEntityType, loggedInUser.UserId),
		IsSavedByUser:    fetchUserSavedStatusByPostId(handlers.saveHelper, postData.ID.Hex(), loggedInUser.UserId),
		PostImpressions:  postImpressionsData[postData.ID],
	}

	postResponse := parsePostResponse(handlers, loggedInUser, postData, &postSecondaryData)

	return postResponse
}

// Internal Method to parse multiple post for response
func parseMultiplePostResponse(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, posts []entities.Post,
) []responses.PostResponse {

	defer utils.Timer("parseMultiplePostResponse")()

	// loggedInUser := LoggedInUserParams{
	// 	UserId:           userId,
	// 	IsCm:             isCm,
	// 	VersionCode:      versionCode,
	// 	PlatformCode:     platformCode,
	// 	ApiRevampCheckV1: apiRevampV1Check,
	// 	MemberRole:       memberRole,
	// }

	postIds := []primitive.ObjectID{}
	for _, post := range posts {
		postIds = append(postIds, post.ID)
	}

	likesCountMap := fetchMultipleEntitiesLikesCount(handlers.likeHelper, postIds, constants.PostEntityType)
	repliesCountMap := fetchMultiplePostsCommentsCount(handlers.commentHelper, postIds)
	repostCountMap := GetRepostCountForMultiplePosts(handlers.widgetHelper, posts)
	isRepostedByUserMap := getIsRepostedByUserForMultiplePosts(handlers.widgetHelper, loggedInUser.UserId, posts)
	isLikedByUserMap := fetchUserLikedStatusForMultipleEntities(handlers.likeHelper, postIds, constants.PostEntityType, loggedInUser.UserId)
	isSavedByUserMap := fetchUserSavedStatusByPostIds(handlers.saveHelper, postIds, loggedInUser.UserId)
	postImpressionsData, _ := computePostImpressionReachCount(handlers, postIds)

	response := []responses.PostResponse{}
	for _, post := range posts {

		postSecondaryData := &PostSecondaryDataParams{
			LikesCount:       likesCountMap[post.ID],
			RepliesCount:     repliesCountMap[post.ID],
			RepostCount:      repostCountMap[post.ID],
			IsRepostedByUser: isRepostedByUserMap[post.ID],
			IsLikedByUser:    isLikedByUserMap[post.ID],
			IsSavedByUser:    isSavedByUserMap[post.ID],
			PostImpressions:  postImpressionsData[post.ID],
		}

		response = append(response, parsePostResponse(handlers, loggedInUser, &post, postSecondaryData))
	}

	return response
}

// Internal Method to fetch post using post_id
func FetchPostData(helper interfaces.PostHelper, postId string, communityId int, isDeletedCheck bool, excludedUserIds []string,
) (*entities.Post, error) {

	// post filter data
	postFilterData := gin.H{
		"_id": postId,
	}

	if communityId > 0 {
		postFilterData["community_id"] = communityId
	}

	if isDeletedCheck {
		postFilterData["is_deleted"] = false
	}

	if len(excludedUserIds) > 0 {
		postFilterData["user_id"] = gin.H{
			"$nin": excludedUserIds,
		}
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

// Internal Method to fetch parsed post with replies
func fetchPostWithReplies(handlers *FeedHandlers, postId string, communityId int, filterOptions map[string]interface{},
	userId string, isCm bool, versionCode string, platformCode string, apiRevampV1Check bool, memberRole string, excludedUserIds []string,
) (responses.PostWithRepliesResponse, error) {

	var postWithRepliesResponse responses.PostWithRepliesResponse

	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, excludedUserIds)
	if err != nil {
		return postWithRepliesResponse, err
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		return postWithRepliesResponse, fmt.Errorf(utils.PostIsHiddenError)
	}

	commentFilterData := gin.H{
		"level":      constants.CommentBaseLevel,
		"is_deleted": false,
		"post_id":    postId,
		"user_id": gin.H{
			"$nin": excludedUserIds,
		},
	}

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, filterOptions)
	if err != nil {
		return postWithRepliesResponse, err
	}

	loggedInUser := LoggedInUserParams{
		UserId:           userId,
		IsCm:             isCm,
		PlatformCode:     platformCode,
		VersionCode:      versionCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	postResponse := parseSinglePostResponse(handlers, postData, &loggedInUser)
	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, userId, isCm, versionCode, platformCode, apiRevampV1Check, handlers.cacheHelper, memberRole)

	postWithRepliesResponse.PostResponse = postResponse
	postWithRepliesResponse.Replies = repliesResponse

	return postWithRepliesResponse, nil
}

// Internal Method to fetch multiple posts data using post_ids
func fetchPostResponseMapFromPostIds(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, postIds []string,
) (map[string]responses.PostResponse, error) {

	// convert post_ids to object ids
	postObjectIds := helpers.ConvertIdsToObjectIds(postIds)

	// filter options to fetch posts from db
	filterOptions := gin.H{
		"_id": gin.H{
			"$in": postObjectIds,
		},
		"community_id": loggedInUser.CommunityId,
	}

	// fetch posts using helper method
	postsLists, err := handlers.postHelper.FindPostHelper(filterOptions, gin.H{})
	if err != nil {
		return nil, err
	}

	// parse fetched posts
	parsedPosts := parseMultiplePostResponse(handlers, loggedInUser, postsLists)

	// Make key value pair of post_id -> PostResponse
	postResponse := map[string]responses.PostResponse{}
	for _, post := range parsedPosts {

		// if post is hidden and user is not cm or creator, then only show isHidden flag
		if !loggedInUser.IsCm && post.IsHidden && loggedInUser.UserId != post.UserId {
			postResponse[post.ID.Hex()] = responses.PostResponse{
				ID:       post.ID,
				IsHidden: true,
			}
		} else {
			postResponse[post.ID.Hex()] = post
		}
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

		activityID, err := handlers.CreateActivity(communityId, []string{userId}, OriginalPostUserID, constants.PostEntity,
			OriginalPostIDObject, OriginalPostUserID, constants.RepostOnPost, ctaData, false, false, primitive.NilObjectID, "")
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
			activityID, err := handlers.CreateActivity(communityId, []string{userId}, member, constants.PostEntity,
				postData.ID, userId, constants.TaggedInPost, ctaData, false, false, primitive.NilObjectID, "")
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

func createPostAfterValidation(handlers *FeedHandlers, userId string, communityId int,
	postRequest *requests.CreatePostRequest, headers map[string]string) (*entities.Post, error) {

	// create post using the helper method
	postId, err := handlers.postHelper.CreatePostHelper(postRequest.Text, postRequest.Heading,
		communityId, userId, postRequest.Attachments, postRequest.ChatroomID,
		postRequest.TempID, postRequest.ParsedTopicIds, postRequest.OriginalAuthor, postRequest.Visibility,
		postRequest.IsRepost, postRequest.IsAnonymous, postRequest.CreatedAt)
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
		updateConnectionList(handlers, userId, communityId, "", false, enums.OneWayConnection)
	}

	userConnectionData, _ = getUserConnectionDataFromCache(handlers, userId, communityId)
	for connectionData := range userConnectionData {
		utils.SafeGo(func() {
			updateConnectionFeedBuffer(handlers, connectionData, communityId, postId.(primitive.ObjectID).Hex(), true)
		})
	}

	// fetch post data using new post_id
	postData, err := FetchPostData(handlers.postHelper, postId.(primitive.ObjectID).Hex(), communityId, true, []string{})
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
func CreatePostAfterValidationFromType(handlers *FeedHandlers, userId string, communityId int, postRequest *requests.CreatePostRequest, headers map[string]string,
) (*entities.Post, error) {

	var postData *entities.Post
	var err error

	// create post based on post type
	if postRequest.PostType == constants.PendingPostEntityType {
		postData, err = createPendingPostAfterValidation(handlers, userId, communityId, postRequest)
	} else {
		postData, err = createPostAfterValidation(handlers, userId, communityId, postRequest, headers)
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
	memberRole := headers[utils.HeadersMemberRole]
	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// bind request body to struct
	var createPostRequest requests.CreatePostRequest
	if err := c.ShouldBindJSON(&createPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             createPostRequest.UserIsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// If users are tagged in post
	if len(createPostRequest.UUIDs) > 0 {
		similarUUIDs := utils.GetSimilarBetweenArray(createPostRequest.UUIDs, blockUserValuesList.BlockedUsers)
		if len(similarUUIDs) > 0 {
			utils.GeneralAPIValidationError(c, utils.BlockedUserTagError)
			return
		}

		similarUUIDs = utils.GetSimilarBetweenArray(createPostRequest.UUIDs, blockUserValuesList.BlockingUsers)
		if len(similarUUIDs) > 0 {
			utils.GeneralAPIValidationError(c, utils.BlockingUserTagError)
			return
		}
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
	postData, err := CreatePostAfterValidationFromType(handlers, userId, communityId, &createPostRequest, headers)
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

	// fetch post response data
	fetchPostData, err := fetchPostWithReplies(handlers, postData.ID.Hex(), communityId, filterOptions, headers[utils.HeadersMemberId], createPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, memberRole, excludedUserIds)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	response := gin.H{
		"post": fetchPostData,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, response)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, createPostRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	utils.GenerateSuccessResponse(c, response)
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
	userId := headers[utils.HeadersMemberId]
	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	var fetchPostQueryRequest requests.FetchPostsQueryRequest

	err := c.BindQuery(&fetchPostQueryRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             fetchPostQueryRequest.UserIsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
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
		postsResponse, err = fetchPostResponseMapFromPostIds(handlers, loggedInUser, postIds)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

	}

	// If pending_post_ids are present, fetch posts data from pending posts
	if len(pendingPostIds) > 0 {

		// Fetch posts data from pending posts using internal method
		pendingPostData, err := fetchMultiplePendingPostsData(handlers, loggedInUser, pendingPostIds)
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
	response["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, response)
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

	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]

	memberRole := headers[utils.HeadersMemberRole]
	userId := headers[utils.HeadersMemberId]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	if paramIsCm == "true" || utils.IsCMRole(memberRole) {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// fetch post response data
	fetchPostData, err := fetchPostWithReplies(handlers, postId, communityId, commentFilterOptions, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, memberRole, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response := gin.H{
		"post": fetchPostData,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, response)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, isCm, headers[utils.HeadersMemberId])

	// return final response
	utils.GenerateSuccessResponse(c, response)
}

// Exposed Method to edit a Post
func (handlers *FeedHandlers) EditPost(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	userId := headers[utils.HeadersMemberId]

	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]

	memberRole := headers[utils.HeadersMemberRole]
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
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
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Check if user is cm or post creator
	if !editPostRequest.UserIsCm && postData.UserId != userId {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             editPostRequest.UserIsCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
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
		errorMeta, err := validateAndUpdatePostImagesForNSFWContent(handlers.cacheHelper, userId, communityId,
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

		editPostRequest.ParsedTopicIds = topicIDs
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, constants.PostEntityType, editPostRequest.Attachments,
		postId, communityId, userId)
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

	postApprovalNeeded := externalHelpers.IsPostApprovalNeeded(handlers.cacheHelper, userId, communityId, editPostRequest.UserIsCm)
	if postApprovalNeeded {
		pendingPostData, err := fetchPendingPostFromPostId(handlers.pendingPostHelper, postData.ID.Hex())
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		if pendingPostData == nil {
			// Create Pending Post with edited data
			createPendingPostFromPost(handlers, postData, userId, communityId, headers, &editPostRequest)

		} else if pendingPostData.Status != enums.UnderReview {
			// Update the Pending Post with edited data and change the status to under review
			err = editPendingPostAfterValidation(handlers, communityId, userId, editPostRequest.Attachments, editPostRequest.Text,
				editPostRequest.Heading, editPostRequest.Visibility, []string{}, pendingPostData,
				true, topicIDs, enums.UnderReview, postId)

			if err != nil {
				utils.GeneralAPIValidationError(c, err.Error())
				return
			}
		} else {
			utils.GeneralAPIValidationError(c, "Post is already in review.")
			return
		}

	} else {
		// Update the post
		_, err = editPostAfterValidation(handlers, communityId, postData.ID, editPostRequest.Text, editPostRequest.Heading,
			updatedAttachments, editPostRequest.TopicIds, existingTopicIds, editPostRequest.Visibility)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post response data
	fetchPostData, err := fetchPostWithReplies(handlers, postId, communityId, commentFilterOptions, userId, editPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, memberRole, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response := gin.H{
		"post": fetchPostData,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, response)
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, editPostRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	utils.GenerateSuccessResponse(c, response)
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
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
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
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
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
		"entity_type": constants.PostEntity,
		"entity_id":   postData.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// delete and fill cache data
	handlers.activityHelper.WarmupUserActivityFeedCache(postData.CommunityId, postData.UserId)
	handlers.activityHelper.WarmupUniversalFeedCache(postData.CommunityId)

	// if deleted by CM, create delete activity
	if deletePostRequest.UserIsCm && headers[utils.HeadersMemberId] != postData.UserId {
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
			postData.UserId, constants.PostEntity, postData.ID, postData.UserId, constants.CMDeletedPost, gin.H{},
			false, false, primitive.NilObjectID, "")
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
	RepostWidgetData := getRepostWidgetDataFromPost(&originalPostData)
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
}

func deleteUserPostRepostActivity(handlers *FeedHandlers, repostPostData *entities.Post, headers map[string]string) error {

	OriginalPostID := repostPostData.Attachments[0].AttachmentMeta.EntityID

	activityFilterData := gin.H{
		"community_id": repostPostData.CommunityId,
		"entity_type":  constants.PostEntity,
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
		"entity_type": constants.CommentEntity,
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
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post using helper method
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	if postData.IsHidden {
		utils.GeneralAPIValidationError(c, utils.PostHiddenCannotPinError)
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
	postData, err = FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
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
	memberRole := headers[utils.HeadersMemberRole]

	if paramIsCm == "true" {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// post filter data
	postFilterData := gin.H{
		"user_id":      userId,
		"is_deleted":   false,
		"community_id": communityId,
	}

	if !isCm {
		postFilterData["$nor"] = []gin.H{
			{
				"is_hidden": true,
				"user_id": gin.H{
					"$ne": userId,
				},
			},
		}
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

	parsedPosts := parseMultiplePostResponse(handlers, loggedInUser, postResults)

	// final response data
	finalResponse := gin.H{
		"success":     true,
		"posts":       parsedPosts,
		"total_count": postsCount,
	}

	finalResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalResponse, communityId)
	finalResponse["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, finalResponse)
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
		updatedPostsWithComments, filtered_comments, err = getTopCommentsAgainstPostsSortOnLikes(handlers, loggedInUser,
			parsedPosts, commentSortOrderVal, universalFeedConfig.CommentCount)

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

func processPostSearchData(handlers *FeedHandlers, loggedInUserParams *LoggedInUserParams, data map[string]interface{},
) []responses.PostResponse {

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

	postResponse := parseMultiplePostResponse(handlers, loggedInUserParams, postList)

	return postResponse
}

// Exposed Method to search Posts
func (handlers *FeedHandlers) SearchPost(c *gin.Context) {
	// fetch query params and headers
	headers := utils.GetHeaders(c)

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	userId := headers[utils.HeadersMemberId]
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]
	memberRole := headers[utils.HeadersMemberRole]

	isCm := utils.IsCMRole(memberRole)

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
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// parsing of chatroom ids
	excludedChatroomIds := utils.ParseIntArrayParam(searchPostRequest.ExcludedChatroomIDs)
	parsedExcludedChatroomIds, _ := json.Marshal(excludedChatroomIds)

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// dsl query to search posts
	postQuery := GetPostFilterQuery(userId, page, pageSize, searchPostRequest.SearchType, searchPostRequest.Search, fmt.Sprintf("%v", string(parsedExcludedChatroomIds)), communityId, excludedUserIds, isCm)
	response := handlers.esHelper.ExecuteQuery(postQuery, constants.PostIndexName)

	finalResponse := processPostSearchData(handlers, loggedInUser, response)

	finalParsedResponse := gin.H{
		"success": true,
		"posts":   finalResponse,
	}

	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, communityId)
	finalParsedResponse["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, finalParsedResponse)
	finalParsedResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalParsedResponse, communityId, searchPostRequest.UserIsCm, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}

// Exposed Method to search user created Posts
func (handlers *FeedHandlers) SearchUserCreatedPost(c *gin.Context) {
	// fetch query params and headers
	userId := c.Param("user_id")
	headers := utils.GetHeaders(c)
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]
	memberRole := headers[utils.HeadersMemberRole]

	isCm := utils.IsCMRole(memberRole)

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
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	if userId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	loggedInUser := &LoggedInUserParams{
		UserId:           userId,
		CommunityId:      communityId,
		IsCm:             isCm,
		VersionCode:      versionCode,
		PlatformCode:     platformCode,
		ApiRevampCheckV1: apiRevampV1Check,
		MemberRole:       memberRole,
	}

	// dsl query to search posts
	postQuery := GetSelfPostFilterQuery(page, pageSize, searchPostRequest.SearchType,
		searchPostRequest.Search, userId, communityId)
	response := handlers.esHelper.ExecuteQuery(postQuery, constants.PostIndexName)

	finalResponse := processPostSearchData(handlers, loggedInUser, response)

	finalParsedResponse := gin.H{
		"success": true,
		"posts":   finalResponse,
	}

	finalParsedResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalParsedResponse, communityId)
	finalParsedResponse["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, finalParsedResponse)
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
func createOrUpdatePostTopics(handlers *FeedHandlers, postId string, deleteAllExisting bool) error {

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

		if len(topicResults) == 0 {
			logging.Error("No topics found for the post")
			return nil
		}

		var parentTopicIds []primitive.ObjectID
		for _, topicResult := range topicResults {
			parentTopicIds = append(parentTopicIds, topicResult.AllParentIds...)
		}

		err = handlers.postTopicsHelper.CreateOrUpdateManyPostTopicsHelper(originalPost.ID, originalPost.TopicIds, parentTopicIds, originalPost.CommunityId)
		if err != nil {
			logging.Error("Error in creating Post Topics: ", err)
			return nil
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

func CreatePostAsyncTasks(handlers *FeedHandlers, postId string) error {

	// Create post topics in PostTopics collection and update post count in topics
	err := createOrUpdatePostTopics(handlers, postId, false)
	if err != nil {
		return err
	}

	// Update personalised feed for the user
	err = updateRecencyMetricForNewlycreatedPost(handlers.postHelper, handlers.cacheHelper, postId)
	if err != nil {
		logging.Error("Error in updating personalised feed for nearly created post: ", err)
		return nil
	}

	return nil
}

func EditPostAsyncTasks(handlers *FeedHandlers, postId string) error {

	// Create post topics in PostTopics collection and update post count in topics
	err := createOrUpdatePostTopics(handlers, postId, true)
	if err != nil {
		return err
	}

	return nil
}

// Exposed method to mark posts as seen
func (handlers *FeedHandlers) MarkPostsSeen(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	isCm := utils.IsCMRole(headers[utils.HeadersMemberRole])

	// Create logged in user params
	loggedInUser := LoggedInUserParams{
		UserId:           headers[utils.HeadersMemberId],
		MemberRole:       headers[utils.HeadersMemberRole],
		CommunityId:      communityId,
		IsCm:             isCm,
		PlatformCode:     headers[utils.HeadersPlatformCode],
		VersionCode:      headers[utils.HeadersVersionCode],
		ApiRevampCheckV1: utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion]),
	}

	var markPostsSeenRequest requests.MarkSeenPostsRequest
	if err := c.ShouldBindJSON(&markPostsSeenRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Validate request
	err := validateMarkPostsSeenRequest(handlers, loggedInUser, markPostsSeenRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Call helper method to mark posts as seen in Background
	utils.SafeGo(func() {
		saveDampenedPostsForUserInDb(handlers, loggedInUser.UserId, loggedInUser.CommunityId, markPostsSeenRequest.PostIds)
	})

	utils.GenerateSuccessResponse(c, nil)
}

func validateMarkPostsSeenRequest(handlers *FeedHandlers, loggedInUser LoggedInUserParams, markPostsSeenRequest requests.MarkSeenPostsRequest) error {

	// Validate post ids
	if len(markPostsSeenRequest.PostIds) == 0 {
		return fmt.Errorf("post IDs are required")
	}

	// validate if post ids are valid
	postIds := helpers.ConvertIdsToObjectIds(markPostsSeenRequest.PostIds)

	// fetch posts using helper method
	postFilterData := gin.H{
		"_id": gin.H{
			"$in": postIds,
		},
		"community_id": loggedInUser.CommunityId,
		"is_deleted":   false,
	}

	posts, err := handlers.postHelper.FindPostHelper(postFilterData, gin.H{})
	if err != nil {
		return err
	}

	if len(posts) != len(postIds) {
		return fmt.Errorf("Invalid post ids sent")
	}

	return nil
}

func fetchPendingPostFromPostId(helper interfaces.PendingPostHelper, postId string) (*entities.PendingPost, error) {
	// filter data
	filterData := gin.H{
		"post_id":    postId,
		"is_deleted": false,
	}

	// fetch post using helper method
	results, err := helper.FindPendingPostHelper(filterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of post_id
	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// Function to create pending post from normal post with updated data
func createPendingPostFromPost(handlers *FeedHandlers, postData *entities.Post, userId string, communityId int, headers map[string]string,
	editPostRequest *requests.EditPostRequest) (*entities.Post, error) {

	cpr := requests.CreatePostRequest{
		Text:           editPostRequest.Text,
		Heading:        editPostRequest.Heading,
		Attachments:    editPostRequest.Attachments,
		ChatroomID:     postData.ChatroomId,
		TopicIds:       editPostRequest.TopicIds,
		ParsedTopicIds: editPostRequest.ParsedTopicIds,
		Visibility:     editPostRequest.Visibility,
		TempID:         postData.TempId,
		IsRepost:       postData.IsRepost,
		PostId:         postData.ID.Hex(),
		PostType:       constants.PendingPostEntityType,
	}

	// create post using internal method
	postDbData, err := CreatePostAfterValidationFromType(handlers, userId, communityId, &cpr, headers)
	if err != nil {
		return nil, err
	}

	return postDbData, nil
}

// Internal method to edit post after validation
func editPostAfterValidation(handlers *FeedHandlers, communityId int, postId primitive.ObjectID, updatedPostText string, updatedPostHeading string,
	updatedAttachments []requests.AttachmentRequest, updatedTopicIds []string, existingTopicIds []primitive.ObjectID,
	updatedPostVisibility string) (*entities.Post, error) {

	// Topic IDs
	topicIds := helpers.ConvertIdsToObjectIds(updatedTopicIds)

	// update post data using helper method
	err := handlers.postHelper.EditPostHelper(postId, updatedPostText, updatedPostHeading, updatedAttachments,
		topicIds, updatedPostVisibility, true)
	if err != nil {
		return nil, err
	}

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId.Hex(), communityId, true, []string{})
	if err != nil {
		return nil, err
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

	if updatedTopicIds != nil {
		updatePostCountInTopics(handlers, updatedTopicIds, existingTopicIds)
	}

	return postData, nil
}

// Exposed Method to update post share count
func (handlers *FeedHandlers) UpdatePostShareCount(c *gin.Context) {
	// fetch headers and url params
	postId := c.Param("post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var updatePostShareCountRequest requests.UpdatePostShareCountRequest
	if err := c.ShouldBindJSON(&updatePostShareCountRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Update the post share count
	if updatePostShareCountRequest.CountNumberType == enums.IncreasePostShareCountType {
		postData.PostShareCount += updatePostShareCountRequest.ShareNumber
	} else if updatePostShareCountRequest.CountNumberType == enums.DecreasePostShareCountType {
		postData.PostShareCount -= updatePostShareCountRequest.ShareNumber
	} else {
		utils.GeneralAPIValidationError(c, "Invalid count number type")
		return
	}

	// update data
	updateData := gin.H{
		"$set": gin.H{
			"post_share_count": postData.PostShareCount,
		},
	}

	// Update the post
	handlers.postHelper.UpdatePostByIdHelper(postData.ID, updateData)

	// fetch updated post data using post_id
	postData, err = FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// insert post data in elastic search
	err = handlers.esHelper.IndexDocument(ParsePostIndexData(postData), postData.ID.Hex(),
		constants.PostIndexName)
	if err != nil {
		logging.Error(fmt.Sprint("Error in updating post data in elastic search: ", err.Error()))
	}

	// return final response
	utils.GenerateSuccessResponse(c, nil)
}

// Exposed Method to hide a Post
func (handlers *FeedHandlers) HidePost(c *gin.Context) {

	// fetch headers
	headers := utils.GetHeaders(c)
	memberRole := headers[utils.HeadersMemberRole]

	// fetch url params
	postId := c.Param("post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// CM Validation
	if !utils.IsCMRole(memberRole) {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// fetch post using helper method
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update data
	updateData := gin.H{
		"$set": gin.H{
			"is_hidden": !postData.IsHidden,
			"is_pinned": false,
		},
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(postData.ID, updateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch updated post data using post_id
	postData, err = FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update post data in elastic search
	err = handlers.esHelper.IndexDocument(ParsePostIndexData(postData), postData.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// return final response
	utils.GenerateSuccessResponse(c, nil)
}
