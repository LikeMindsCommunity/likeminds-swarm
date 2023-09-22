package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to process poll attachment data
func processPollCustomAttachmentData(metaData map[string]interface{}) map[string]interface{} {
	if _, exists := metaData["is_anonymous"]; !exists {
		metaData["is_anonymous"] = false
	}

	if _, exists := metaData["allow_add_option"]; !exists {
		metaData["allow_add_option"] = false
	}

	if _, exists := metaData["poll_type"]; !exists {
		metaData["poll_type"] = enums.InstantPollType
	}

	if _, exists := metaData["multiple_select_state"]; !exists {
		metaData["multiple_select_state"] = enums.ExactlySelectStateType
	}

	if _, exists := metaData["multiple_select_number"]; !exists {
		metaData["multiple_select_number"] = 1
	}

	return metaData
}

// Internal Method to process meta data before widget creation
func processMetaBeforeWidgetCreation(attachment requests.Attachment, metaData map[string]interface{},
	lmMeta map[string]interface{}, uuid string) (map[string]interface{}, map[string]interface{}, bool) {
	switch attachment.AttachmentType {
	case enums.PollWidget:
		// create poll options
		pollOptionObjects, err := createPollOptionObjects(attachment.AttachmentMeta.Options, uuid)
		if err != nil {
			return metaData, lmMeta, false
		}

		lmMeta["options"] = pollOptionObjects
		delete(metaData, "options")

	default:
		if len(lmMeta) == 0 {
			lmMeta = nil
		}
	}

	return metaData, lmMeta, true
}

// Internal Method to process meta data before widget edition
func processMetaBeforeWidgetEdition(attachment requests.Attachment, metaData map[string]interface{},
	existingMetaData map[string]interface{}) map[string]interface{} {
	updatedMetaData := existingMetaData

	if attachment.AttachmentType == enums.PollWidget {
		if _, exists := metaData["title"]; exists {
			updatedMetaData["title"] = metaData["title"]
		}
	} else {
		updatedMetaData = metaData
	}

	delete(updatedMetaData, "entity_id")

	return updatedMetaData
}

// Internal Method to process attachments for widgets
func processAttachmentsForWidgets(c *gin.Context, handlers *FeedHandlers, attachments []requests.Attachment,
	postId string, communityId int, uuid string) ([]requests.Attachment, bool) {
	// process attachments for custom widgets
	updatedAttachments := []requests.Attachment{}

	for _, attachment := range attachments {
		isLMCreatedCustomWidget := false

		switch attachment.AttachmentType {
		case enums.PollWidget, enums.ArticleWidget:
			isLMCreatedCustomWidget = true
		}

		switch attachment.Type {
		case enums.PollType, enums.ArticleType:
			isLMCreatedCustomWidget = true
		}

		if isLMCreatedCustomWidget {
			// meta data conversion to desired type
			metaData := map[string]interface{}{}
			entityId := ""

			convertedMetaData, _ := json.Marshal(attachment.AttachmentMeta)
			_ = json.Unmarshal(convertedMetaData, &metaData)

			switch attachment.AttachmentType {
			case enums.PollWidget:
				metaData = processPollCustomAttachmentData(metaData)
			}

			// Edit the metadata keys in case entity_id already exists in LM Created widget
			if attachment.AttachmentMeta.EntityID != "" {
				widgetData, err := fetchWidgetByID(handlers.widgetHelper, attachment.AttachmentMeta.EntityID, true, communityId)
				if err != nil {
					return nil, false
				}

				// process meta data before widget edition
				updatedMetaData := processMetaBeforeWidgetEdition(attachment, metaData, widgetData.MetaData)

				// update widget from given metadata
				_, ok := editWidget(c, handlers, attachment.AttachmentMeta.EntityID, true, updatedMetaData, nil, communityId)
				if !ok {
					return nil, false
				}

				entityId = attachment.AttachmentMeta.EntityID

				// Else create a new LM Created widget
			} else {
				// Generate LM Meta
				lmMeta := map[string]interface{}{}

				// process meta data before widget creation
				metaData, lmMeta, ok := processMetaBeforeWidgetCreation(attachment, metaData, lmMeta, uuid)
				if !ok {
					return nil, false
				}

				// create widget from given metadata
				widgetData, ok := createWidget(c, handlers, true, postId, constants.PostEntityType, metaData, lmMeta, communityId)
				if !ok {
					return nil, false
				}

				entityId = widgetData.ID.Hex()

			}

			// updated attachment
			updatedAttachment := requests.Attachment{
				AttachmentType: attachment.AttachmentType,
				AttachmentMeta: requests.AttachmentMeta{
					EntityID: entityId,
				},
			}

			updatedAttachments = append(updatedAttachments, updatedAttachment)

			// Else do nothing
		} else {
			updatedAttachments = append(updatedAttachments, attachment)
		}
	}

	return updatedAttachments, true
}

// Internal Method to parse Post Attachments
func parsePostAttachments(attachments []entities.Attachment, versionCode string,
	platformCode string, apiRevampV1Check bool) []entities.Attachment {
	parsedAttachments := []entities.Attachment{}
	feedLinkMediaCheck := utils.CheckVersionInverted(utils.FeedLinkMediaVersion, versionCode, platformCode)
	feedVideoAndDocumentMediaCheck := utils.CheckVersionInverted(utils.FeedVideoAndDocumentMediaVersions, versionCode, platformCode)
	newAttachmentMeta := entities.AttachmentMeta{
		Url: constants.AttachmentNotFoundImageUrl,
	}

	for _, attachment := range attachments {
		showUpdateAppImage := false

		if feedLinkMediaCheck && attachment.AttachmentType == enums.LinkWidget {
			showUpdateAppImage = true
		}

		if feedVideoAndDocumentMediaCheck && (attachment.AttachmentType == enums.VideoWidget || attachment.AttachmentType == enums.DocumentWidget) {
			showUpdateAppImage = true
		}

		if showUpdateAppImage {
			attachment.AttachmentType = enums.ImageWidget
			attachment.AttachmentMeta = &newAttachmentMeta
		}

		parsedAttachments = append(parsedAttachments, attachment)
	}

	// Api revamp check for attachments
	if apiRevampV1Check {
		for i := range parsedAttachments {

			// Update attachment_type from type and remove attachment_type
			parsedAttachments[i].Type = enums.NewAttachmentTypeFromInt(parsedAttachments[i].AttachmentType)
			parsedAttachments[i].AttachmentType = 0

			// Update attachment_meta from meta_data and remove attachment_meta
			parsedAttachments[i].MetaData = parsedAttachments[i].AttachmentMeta
			parsedAttachments[i].AttachmentMeta = nil
		}
	}

	return parsedAttachments
}

// Internal Method to validate image attachment
func validateImageAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.Url == "" {
		return "send url in attachment_meta for image", false
	}

	return "", true
}

// Internal Method to validate video attachment
func validateVideoAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.Url == "" {
		return "send url in attachment_meta for video", false
	}

	if attachment.AttachmentMeta.Duration == 0 {
		return "send duration in attachment_meta for video", false
	}

	return "", true
}

// Internal Method to validate document attachment
func validateDocumentAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.Url == "" {
		return "send url in attachment_meta for document", false
	}

	if attachment.AttachmentMeta.Format == "" {
		return "send format in attachment_meta for document", false
	}

	if attachment.AttachmentMeta.Size == 0 {
		return "send size in attachment_meta for document", false
	}

	return "", true
}

// Internal Method to validate link attachment
func validateLinkAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.OgTags.Url == "" {
		return "send url in og_tags in attachment_meta for link", false
	}

	return "", true
}

// Internal Method to validate custom attachment
func validateCustomAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.EntityID == "" {
		return "send entity_id in attachment_meta for custom widget", false
	}

	return "", true
}

// Internal Method to validate poll attachment
func validatePollAttachment(attachment requests.Attachment, isEditRequest bool) (string, bool) {
	if attachment.AttachmentMeta.Title == "" {
		return "send title in attachment_meta for poll widget", false
	}

	if !isEditRequest {
		if len(attachment.AttachmentMeta.Options) == 0 {
			return "send options in attachment_meta for poll widget", false
		}

		if attachment.AttachmentMeta.PollType != "" && !enums.IsPollTypeValid(attachment.AttachmentMeta.PollType) {
			return "send valid poll_type in attachment_meta for poll widget", false
		}

		if attachment.AttachmentMeta.MultipleSelectState != "" && !enums.IsPollMultipleSelectStateValid(attachment.AttachmentMeta.MultipleSelectState) {
			return "send valid multiple_select_state in attachment_meta for poll widget", false
		}

		if attachment.AttachmentMeta.MultipleSelectNumber < 0 {
			return "Send valid multiple_select_number in attachment_meta for poll widget", false
		}

		if (attachment.AttachmentMeta.ExpiryTime == 0) ||
			(attachment.AttachmentMeta.ExpiryTime != 0 && attachment.AttachmentMeta.ExpiryTime <= int64(time.Now().UnixMilli())) {
			return "Send valid expiry_time in attachment_meta for poll widget", false
		}
	}

	return "", true
}

// Internal Method to validate article attachment
func validateArticleAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.Body == "" {
		return "Send body in attachment_meta for article", false
	}

	if attachment.AttachmentMeta.Title == "" {
		return "Send title in attachment_meta for article", false
	}

	if attachment.AttachmentMeta.CoverImageUrl == "" {
		return "Send cover_image_url in attachment_meta for article", false
	}

	return "", true
}

// Internal method to validate attachments for post
func validateAndUpdatePostAttachments(c *gin.Context, attachments []requests.Attachment, apiRevampV1check bool,
	isEditRequest bool) bool {

	// Api revamp check to validate and update attachments
	if apiRevampV1check {

		for i := range attachments {

			// If type in attachments is not empty
			if attachments[i].Type != "" {

				// Check if attachment type is valid
				if !attachments[i].Type.IsValid() {
					utils.GeneralAPIValidationError(c, "Invalid attachment type: "+attachments[i].Type.ToString())
					return false
				}

				// Update attachment_type from type
				attachments[i].AttachmentType = attachments[i].Type.ToInt()

				// Update attachment_meta from meta_data
				attachments[i].AttachmentMeta = attachments[i].MetaData
			}

			// validate attachment urls if present
			urlArray := []string{
				attachments[i].AttachmentMeta.Url,
				attachments[i].AttachmentMeta.ThumbnailUrl,
				attachments[i].AttachmentMeta.OgTags.Url,
				attachments[i].AttachmentMeta.CoverImageUrl,
			}

			err := helpers.AreValidURLs(urlArray)
			if err != "" {
				utils.GeneralAPIValidationError(c, err)
				return false
			}
		}

	}

	// validate attachment_meta
	for _, element := range attachments {
		switch element.AttachmentType {
		case enums.ImageWidget:
			errorMessage, ok := validateImageAttachment(element)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		case enums.VideoWidget:
			errorMessage, ok := validateVideoAttachment(element)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		case enums.DocumentWidget:
			errorMessage, ok := validateDocumentAttachment(element)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		case enums.LinkWidget:
			errorMessage, ok := validateLinkAttachment(element)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		case enums.CustomWidget:
			errorMessage, ok := validateCustomAttachment(element)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		case enums.PollWidget:
			errorMessage, ok := validatePollAttachment(element, isEditRequest)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		case enums.ArticleWidget:
			errorMessage, ok := validateArticleAttachment(element)
			if !ok {
				utils.GeneralAPIValidationError(c, errorMessage)
				return false
			}

		default:
			utils.GeneralAPIValidationError(c, "send valid attachment_type in attachment")
			return false
		}
	}

	return true
}

// Internal Method to parse response for fetch multiple posts api
func parseFetchMultiplePostResponse(postHelper interfaces.PostHelper, posts []requests.PostResponse, posts_count int64) requests.FetchUserMultiplePostResponse {
	response := requests.FetchUserMultiplePostResponse{}

	response.Success = true
	response.Posts = posts

	if posts_count > 0 {
		response.TotalCount = int(posts_count)
	}

	return response
}

// Internal Method to parse topics response
func parseTopicsResponse(topicHelper interfaces.TopicHelper, topicIds []primitive.ObjectID, communityId int) (map[string]requests.TopicResponse, error) {
	// Fetch topics using topic Ids
	topics, err := fetchTopicsByIDs(topicHelper, topicIds, communityId, false)
	if err != nil {
		return nil, err
	}

	topicsResponse := map[string]requests.TopicResponse{}

	// Parse all fetched topics Data
	for _, topic := range topics {
		topicsResponse[topic.ID.Hex()] = parseTopicResponse(&topic)
	}

	return topicsResponse, nil
}

// Internal Method to parse widgets response
func parseWidgetsResponse(handlers *FeedHandlers, widgetIds []primitive.ObjectID, communityId int, uuid string) (map[string]requests.WidgetResponse, error) {
	// Fetch widgets using widget Ids
	widgets, err := fetchWidgetsByIDs(handlers.widgetHelper, widgetIds, communityId)
	if err != nil {
		return nil, err
	}

	widgetsResponse := map[string]requests.WidgetResponse{}

	// Parse all fetched widgets Data
	for _, widget := range widgets {
		widgetsResponse[widget.ID.Hex()] = parseWidgetResponse(handlers, &widget, communityId, uuid)
	}

	return widgetsResponse, nil
}

// Internal Method to parse topic_ids from posts
func getTopicIdsFromPosts(response interface{}) []primitive.ObjectID {
	uniqueTopicIds := []primitive.ObjectID{}
	tempTopicIds := map[primitive.ObjectID]bool{}

	if post, ok := response.(gin.H)["post"]; ok {
		for _, topicId := range post.(requests.FetchPostResponse).Topics {
			if _, exists := tempTopicIds[topicId]; !exists {
				tempTopicIds[topicId] = true
			}
		}
	}

	if posts, ok := response.(gin.H)["posts"]; ok {
		for _, post := range posts.([]requests.PostResponse) {
			for _, topicId := range post.Topics {
				if _, exists := tempTopicIds[topicId]; !exists {
					tempTopicIds[topicId] = true
				}
			}
		}
	}

	for key := range tempTopicIds {
		uniqueTopicIds = append(uniqueTopicIds, key)
	}

	return uniqueTopicIds
}

// Internal Method to parse widget_ids from attachments
func getWidgetIdsFromAttachments(attachments []entities.Attachment) []primitive.ObjectID {
	widgetIds := map[primitive.ObjectID]bool{}
	finalWidgetIds := []primitive.ObjectID{}

	for _, attachment := range attachments {
		entityId := primitive.NilObjectID
		if attachment.AttachmentMeta != nil {
			entityId = attachment.AttachmentMeta.EntityID
		} else if attachment.MetaData != nil {
			entityId = attachment.MetaData.EntityID
		}

		if entityId != primitive.NilObjectID {
			if _, exists := widgetIds[entityId]; !exists {
				widgetIds[entityId] = true
			}
		}
	}

	for key := range widgetIds {
		finalWidgetIds = append(finalWidgetIds, key)
	}

	return finalWidgetIds
}

// Internal Method to parse widget_ids from posts
func getWidgetIdsFromPosts(response interface{}) []primitive.ObjectID {
	uniqueWidgetIds := []primitive.ObjectID{}
	tempWidgetIds := map[primitive.ObjectID]bool{}

	if post, ok := response.(gin.H)["post"]; ok {
		widgetIds := getWidgetIdsFromAttachments(post.(requests.FetchPostResponse).Attachments)

		for _, widgetId := range widgetIds {
			if _, exists := tempWidgetIds[widgetId]; !exists {
				tempWidgetIds[widgetId] = true
			}
		}
	}

	if posts, ok := response.(gin.H)["posts"]; ok {
		for _, post := range posts.([]requests.PostResponse) {
			widgetIds := getWidgetIdsFromAttachments(post.Attachments)

			for _, widgetId := range widgetIds {
				if _, exists := tempWidgetIds[widgetId]; !exists {
					tempWidgetIds[widgetId] = true
				}
			}
		}
	}

	for key := range tempWidgetIds {
		uniqueWidgetIds = append(uniqueWidgetIds, key)
	}

	return uniqueWidgetIds
}

// Internal Method to get topics Data from Posts response
func getTopicDataFromPosts(topicHelper interfaces.TopicHelper, response interface{}, communityId int) map[string]requests.TopicResponse {
	topicIds := getTopicIdsFromPosts(response)

	topicsData, _ := parseTopicsResponse(topicHelper, topicIds, communityId)

	return topicsData
}

// Internal Method to get widget Data from Posts response
func getWidgetDataFromPosts(handlers *FeedHandlers, response interface{}, communityId int, uuid string) map[string]requests.WidgetResponse {
	widgetIds := getWidgetIdsFromPosts(response)

	widgetsData, _ := parseWidgetsResponse(handlers, widgetIds, communityId, uuid)

	return widgetsData
}

// Internal Method to parse post for response
func parsePostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	saveHelper interfaces.SaveHelper, topicHelper interfaces.TopicHelper, post entities.Post,
	userId string, isCm bool, versionCode string, platformCode string, apiRevampV1Check bool) requests.PostResponse {
	likes_count, _ := fetchEntityLikesCount(likeHelper, post.ID.Hex(), constants.PostEntityType)
	replies_count, _ := fetchPostCommentsCount(commentHelper, post.ID.Hex())

	var response requests.PostResponse

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
	response.Attachments = parsePostAttachments(post.Attachments, versionCode, platformCode, apiRevampV1Check)
	response.LikesCount = int(likes_count)
	response.CommentsCount = int(replies_count)
	response.IsDeleted = post.IsDeleted
	response.IsEdited = post.IsEdited
	response.IsLiked = fetchUserLikedStatusByEntity(likeHelper, post.ID.Hex(),
		constants.PostEntityType, userId)
	response.IsSaved = fetchUserSavedStatusByPostId(saveHelper, post.ID.Hex(), userId)
	response.MenuItems = getEntityMenuItems(constants.PostEntityType, isCm,
		userId == post.UserId, post.IsPinned, versionCode, platformCode)

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

	return response
}

// Internal Method to parse multiple post for response
func parseMultiplePostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	saveHelper interfaces.SaveHelper, topicHelper interfaces.TopicHelper, posts []entities.Post, userId string,
	isCm bool, versionCode string, platformCode string, apiRevampV1Check bool) []requests.PostResponse {
	response := []requests.PostResponse{}

	for _, post := range posts {
		response = append(response, parsePostResponse(likeHelper, commentHelper, saveHelper, topicHelper,
			post, userId, isCm, versionCode, platformCode, apiRevampV1Check))
	}

	return response
}

// Internal Method to parse response for fetch post api
func parseFetchPostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	parsedPost requests.PostResponse, replies []requests.CommentResponse) requests.FetchPostResponse {
	var response requests.FetchPostResponse

	response.PostResponse = parsedPost

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []requests.CommentResponse{}
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

// Internal Method to fetch post data
func fetchPostData(handlers *FeedHandlers, postId string, communityId int,
	filterOptions map[string]interface{}, memberId string, isCm bool, versionCode string,
	platformCode string, apiRevampV1Check bool) (interface{}, error) {
	postData, err := fetchPost(handlers.postHelper, postId, communityId)
	if err != nil {
		return nil, err
	}

	commentFilterData := gin.H{
		"level":      constants.CommentBaseLevel,
		"is_deleted": false,
		"post_id":    postId,
	}

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, filterOptions)
	if err != nil {
		return nil, err
	}

	postResponse := parsePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, handlers.topicHelper, *postData, memberId, isCm, versionCode, platformCode,
		apiRevampV1Check)
	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentResults, memberId, isCm, versionCode, platformCode, apiRevampV1Check)
	fetchPostResponse := parseFetchPostResponse(handlers.likeHelper, handlers.commentHelper,
		postResponse, repliesResponse)

	return fetchPostResponse, nil
}

// Internal Method to fetch multiple posts data using post_ids
func fetchMultiplePostsData(handlers *FeedHandlers, postIds []string, communityId int, userId string,
	isCm bool, versionCode string, platformCode string,
	apiRevampV1Check bool) (map[string]requests.PostResponse, error) {

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
	postResponse := map[string]requests.PostResponse{}

	// parse post response data for each post
	for _, post := range postsLists {
		postResponse[post.ID.Hex()] = parsePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper,
			handlers.topicHelper, post, userId, isCm, versionCode, platformCode, apiRevampV1Check)
	}

	return postResponse, nil

}

// Exposed Method to create a Post
func (handlers *FeedHandlers) CreatePost(c *gin.Context) {
	// fetch headers
	headers := utils.GetHeaders(c)

	// Post owner user_id
	postUserId := headers[utils.HeadersMemberId]

	// Set OriginalAuthorUUID to empty string for new posts
	OriginalAuthorUUID := ""

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createPostRequest requests.CreatePostRequest
	if err := c.ShouldBindJSON(&createPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	UserIsCM := createPostRequest.User_is_cm

	// strip text to check if it is empty
	createPostRequest.Text = strings.Trim(createPostRequest.Text, " ")

	if createPostRequest.Text == "" && len(createPostRequest.Attachments) == 0 {
		utils.GeneralAPIValidationError(c, "can't create post without content")
		return
	}

	// validation of attachments
	success := validateAndUpdatePostAttachments(c, createPostRequest.Attachments, apiRevampV1Check, false)
	if !success {
		return
	}

	// convert topic_ids to object ids
	topicIDs := helpers.ConvertIdsToObjectIds(createPostRequest.TopicIds)

	// fetch all the topics sent in the create post body
	if len(topicIDs) > 0 {
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

	// if on_behalf_of_uuid is not empty
	if createPostRequest.On_behalf_of_uuid != "" {
		// Validate if user is cm or not
		if !UserIsCM {
			utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
			return
		}

		// update postUserId and OriginalAuthorUUID
		OriginalAuthorUUID = postUserId
		postUserId = createPostRequest.On_behalf_of_uuid
	}

	// create post using the helper method
	postId, err := handlers.postHelper.CreatePostHelper(createPostRequest.Text, createPostRequest.Heading,
		communityId, postUserId, createPostRequest.Attachments, createPostRequest.ChatroomID,
		createPostRequest.TempID, topicIDs, OriginalAuthorUUID)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// process attachments for widgets
	updatedAttachments, ok := processAttachmentsForWidgets(c, handlers, createPostRequest.Attachments,
		postId.(primitive.ObjectID).Hex(), communityId, postUserId)
	if !ok {
		return
	}

	// update post data using helper method
	err = handlers.postHelper.EditPostHelper(postId.(primitive.ObjectID), createPostRequest.Text,
		createPostRequest.Heading, updatedAttachments, topicIDs, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch post data using new post_id
	postData, err := fetchPost(handlers.postHelper, postId.(primitive.ObjectID).Hex(), communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// insert post data in elastic search
	err = handlers.esHelper.InsertDocument(c, ParsePostIndexData(postData), postData.ID.Hex(),
		constants.PostIndexName)
	if err != nil {
		log.Error(err.Error())
	}

	// Get tagged members from request
	taggedMembers := createPostRequest.UUIDs

	for _, member := range taggedMembers {
		// create tag activity
		activityID, err := handlers.CreateActivity(communityId, []string{postUserId}, member, constants.Post, postId.(primitive.ObjectID), postUserId,
			constants.TaggedInPost, gin.H{
				"entity_type": constants.PostEntityType,
				"post_id":     postId.(primitive.ObjectID).Hex(),
			}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
		}

	}

	// filter options
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
	fetchPostData, err := fetchPostData(handlers, postId.(primitive.ObjectID).Hex(), communityId,
		filterOptions, headers[utils.HeadersMemberId], UserIsCM, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check)
	if err == nil {
		response["post"] = fetchPostData
		response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
		response["widgets"] = getWidgetDataFromPosts(handlers, response, communityId, headers[utils.HeadersMemberId])
	}

	// return final response
	c.JSON(http.StatusOK, response)
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

	// Get Query Params
	paramPostIds := c.Query("post_ids")
	paramIsCm, _ := strconv.ParseBool(c.Query("user_is_cm"))

	// If user is not cm, return error
	if !paramIsCm {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Unmarshal post_ids
	var postIds []string
	err := json.Unmarshal([]byte(paramPostIds), &postIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch multiple posts data using internal method
	postsResponse, err := fetchMultiplePostsData(handlers, postIds, communityId, headers[utils.HeadersMemberId],
		true, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"posts":   postsResponse,
		"success": true,
	}

	postsList := []requests.PostResponse{}
	for _, value := range postsResponse {
		postsList = append(postsList, value)
	}

	parsedResponse := gin.H{
		"posts": postsList,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, parsedResponse, communityId)
	response["widgets"] = getWidgetDataFromPosts(handlers, parsedResponse, communityId, headers[utils.HeadersMemberId])

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
	fetchPostData, err := fetchPostData(handlers, postId, communityId, commentFilterOptions,
		headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}
	response["post"] = fetchPostData
	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["widgets"] = getWidgetDataFromPosts(handlers, response, communityId, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to edit a Post
func (handlers *FeedHandlers) EditPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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
	success := validateAndUpdatePostAttachments(c, editPostRequest.Attachments, apiRevampV1Check, true)
	if !success {
		return
	}

	// strip text and check if it is empty
	editPostRequest.Text = strings.TrimSpace(editPostRequest.Text)

	if editPostRequest.Text == "" && len(editPostRequest.Attachments) == 0 {
		utils.GeneralAPIValidationError(c, "Can't Edit post without content")
		return
	}

	topicIDs := postData.TopicIds

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
	updatedAttachments, ok := processAttachmentsForWidgets(c, handlers, editPostRequest.Attachments, postId, communityId,
		headers[utils.HeadersMemberId])
	if !ok {
		return
	}

	// update post data using helper method
	err = handlers.postHelper.EditPostHelper(postData.ID, editPostRequest.Text, editPostRequest.Heading, updatedAttachments,
		topicIDs, true)
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
	fetchPostData, err := fetchPostData(handlers, postId, communityId, commentFilterOptions, headers[utils.HeadersMemberId],
		editPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode],
		apiRevampV1Check)
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

	// update post data in elastic search
	err = handlers.esHelper.UpdateDocument(c, ParsePostIndexData(postData), postData.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	response := gin.H{
		"success": true,
		"post":    fetchPostData,
	}

	response["topics"] = getTopicDataFromPosts(handlers.topicHelper, response, communityId)
	response["widgets"] = getWidgetDataFromPosts(handlers, response, communityId, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)

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

	// delete post data in elastic search
	err = handlers.esHelper.DeleteDocument(c, postData.ID.Hex(), constants.PostIndexName)
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
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, postData.UserId, constants.Post, postData.ID, postData.UserId, constants.CMDeletedPost, gin.H{}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		if activityID != nil {
			SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
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

	postCommentIds := [](primitive.ObjectID){}

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

	// update post data in elastic search
	err = handlers.esHelper.UpdateDocument(c, ParsePostIndexData(postData), postData.ID.Hex(),
		constants.PostIndexName)
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

	createdPostResponse := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, handlers.topicHelper, postResults, userId, isCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)

	response := parseFetchMultiplePostResponse(handlers.postHelper, createdPostResponse, postsCount)

	// response data
	finalResponse := gin.H{
		"posts":   response.Posts,
		"success": response.Success,
	}

	if response.TotalCount > 0 {
		finalResponse["total_count"] = response.TotalCount
	}

	finalResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalResponse, communityId)
	finalResponse["widgets"] = getWidgetDataFromPosts(handlers, finalResponse, communityId, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, finalResponse)
}

func processPostSearchData(handlers *FeedHandlers, data map[string]interface{}, userId string,
	isCm bool, versionCode string, platformCode string, apiRevampV1Check bool) []requests.PostResponse {
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

	postResponse := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, handlers.topicHelper, postList, userId, isCm, versionCode, platformCode,
		apiRevampV1Check)

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
	excludedChatroomIds := parseIntArrayParam(searchPostRequest.ExcludedChatroomIDs)
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
	finalParsedResponse["widgets"] = getWidgetDataFromPosts(handlers, finalParsedResponse, communityId, headers[utils.HeadersMemberId])

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
	finalParsedResponse["widgets"] = getWidgetDataFromPosts(handlers, finalParsedResponse, communityId, headers[utils.HeadersMemberId])

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
