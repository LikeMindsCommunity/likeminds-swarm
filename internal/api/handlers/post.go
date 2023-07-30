package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse Post Attachments
func parsePostAttachments(attachments []entities.Attachment, versionCode string,
	platformCode string) []entities.Attachment {
	parsedAttachments := []entities.Attachment{}
	feedLinkMediaCheck := utils.CheckVersionInverted(utils.FeedLinkMediaVersion, versionCode, platformCode)
	feedVideoAndDocumentMediaCheck := utils.CheckVersionInverted(utils.FeedVideoAndDocumentMediaVersions, versionCode, platformCode)
	newAttachmentMeta := entities.AttachmentMeta{
		Url: constants.AttachmentNotFoundImageUrl,
	}

	for _, attachment := range attachments {
		showUpdateAppImage := false

		if feedLinkMediaCheck && attachment.AttachmentType == constants.LinkWidget {
			showUpdateAppImage = true
		}

		if feedVideoAndDocumentMediaCheck && (attachment.AttachmentType == constants.VideoWidget || attachment.AttachmentType == constants.DocumentWidget) {
			showUpdateAppImage = true
		}

		if showUpdateAppImage {
			attachment.AttachmentType = constants.ImageWidget
			attachment.AttachmentMeta = &newAttachmentMeta
		}

		parsedAttachments = append(parsedAttachments, attachment)
	}

	return parsedAttachments
}

// Internal method to validate attachments for post
func validatePostAttachments(c *gin.Context, attachments []requests.Attachment) bool {

	for _, element := range attachments {
		switch element.AttachmentType {
		case constants.ImageWidget:
			if element.AttachmentMeta.Url == "" {
				utils.GeneralAPIValidationError(c, "send url in attachment_meta for image")
				return false
			}

		case constants.VideoWidget:
			if element.AttachmentMeta.Url == "" {
				utils.GeneralAPIValidationError(c, "send url in attachment_meta for video")
				return false
			}

			if element.AttachmentMeta.Duration == 0 {
				utils.GeneralAPIValidationError(c, "send duration in attachment_meta for video")
				return false
			}

		case constants.DocumentWidget:
			if element.AttachmentMeta.Url == "" {
				utils.GeneralAPIValidationError(c, "send url in attachment_meta for document")
				return false
			}

			if element.AttachmentMeta.Format == "" {
				utils.GeneralAPIValidationError(c, "send format in attachment_meta for document")
				return false
			}

			if element.AttachmentMeta.Size == 0 {
				utils.GeneralAPIValidationError(c, "send size in attachment_meta for document")
				return false
			}

		case constants.LinkWidget:
			if element.AttachmentMeta.OgTags.Url == "" {
				utils.GeneralAPIValidationError(c, "send url in og_tags in attachment_meta for link")
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
func parseFetchMultiplePostResponse(postHelper interfaces.PostHelper, posts []requests.PostResponse,
	posts_count int64) requests.FetchUserMultiplePostResponse {
	response := requests.FetchUserMultiplePostResponse{}

	response.Success = true
	response.Posts = posts

	if posts_count > 0 {
		response.TotalCount = int(posts_count)
	}

	return response
}

// Internal Method to parse topics response
func parseTopicsResponse(topicHelper interfaces.TopicHelper, topicIds []primitive.ObjectID, communityId int) ([]requests.TopicResponse, error) {
	// Fetch topics using topic Ids
	topics, err := fetchTopicsByIDs(topicHelper, topicIds, communityId, false)
	if err != nil {
		return nil, err
	}

	topicsResponse := []requests.TopicResponse{}

	// Parse all fetched topics Data
	for _, topic := range topics {
		topicsResponse = append(topicsResponse, parseTopicResponse(&topic))
	}

	return topicsResponse, nil
}

// Internal Method to parse post for response
func parsePostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	saveHelper interfaces.SaveHelper, topicHelper interfaces.TopicHelper, post entities.Post,
	userId string, isCm bool, versionCode string, platformCode string) requests.PostResponse {
	likes_count, _ := fetchEntityLikesCount(likeHelper, post.ID.Hex(), constants.PostEntityType)
	replies_count, _ := fetchPostCommentsCount(commentHelper, post.ID.Hex())
	topics, _ := parseTopicsResponse(topicHelper, post.TopicIds, post.CommunityId)

	var response requests.PostResponse

	response.ID = post.ID
	response.TempID = post.TempId
	response.Text = post.Text
	response.Topics = topics
	response.Heading = post.Heading
	response.CommunityId = post.CommunityId
	response.ChatroomId = post.ChatroomId
	response.IsPinned = post.IsPinned
	response.UserId = post.UserId
	response.UUID = post.UserId
	response.Attachments = parsePostAttachments(post.Attachments, versionCode, platformCode)
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

	return response
}

// Internal Method to parse multiple post for response
func parseMultiplePostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	saveHelper interfaces.SaveHelper, topicHelper interfaces.TopicHelper, posts []entities.Post, userId string,
	isCm bool, versionCode string, platformCode string) []requests.PostResponse {
	response := []requests.PostResponse{}

	for _, post := range posts {
		response = append(response, parsePostResponse(likeHelper, commentHelper, saveHelper, topicHelper,
			post, userId, isCm, versionCode, platformCode))
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
	platformCode string) (interface{}, error) {
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
		handlers.saveHelper, handlers.topicHelper, *postData, memberId, isCm, versionCode, platformCode)
	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentResults, memberId, isCm, versionCode, platformCode)
	fetchPostResponse := parseFetchPostResponse(handlers.likeHelper, handlers.commentHelper,
		postResponse, repliesResponse)

	return fetchPostResponse, nil
}

// Internal Method to fetch multiple posts data using post_ids
func fetchMultiplePostsData(handlers *FeedHandlers,
	postIds []string,
	communityId int,
	userId string,
	isCm bool,
	versionCode string,
	platformCode string) (map[string]requests.PostResponse, error) {

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
			handlers.topicHelper, post, userId, isCm, versionCode, platformCode)
	}

	return postResponse, nil

}

// Exposed Method to create a Post
func (handlers *FeedHandlers) CreatePost(c *gin.Context) {
	// fetch headers
	headers := utils.GetHeaders(c)

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

	// strip text to check if it is empty
	createPostRequest.Text = strings.Trim(createPostRequest.Text, " ")

	if createPostRequest.Text == "" && len(createPostRequest.Attachments) == 0 {
		utils.GeneralAPIValidationError(c, "can't create post without content")
		return
	}

	// validation of attachment objects
	success := validatePostAttachments(c, createPostRequest.Attachments)
	if !success {
		return
	}

	// convert topic_ids to object ids
	topicIDs := helpers.ConvertIdsToObjectIds(createPostRequest.TopicIds)

	// fetch all the topics sent in the create post body
	if len(topicIDs) > 0 {
		topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, true)
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

	// create post using the helper method
	postId, err := handlers.postHelper.CreatePostHelper(createPostRequest.Text,
		createPostRequest.Heading, communityId, headers[utils.HeadersMemberId],
		createPostRequest.Attachments, createPostRequest.ChatroomID, createPostRequest.TempID,
		topicIDs)
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
		activityID, err := handlers.CreateActivity(communityId, []string{headers[utils.HeadersMemberId]}, member, constants.Post, postId.(primitive.ObjectID), headers[utils.HeadersMemberId], constants.TaggedInPost, gin.H{
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
		filterOptions, headers[utils.HeadersMemberId], false, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode])
	if err == nil {
		response["post"] = fetchPostData
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to fetch multiple posts from post_ids
func (handlers *FeedHandlers) FetchPosts(c *gin.Context) {

	// fetch headers
	headers := utils.GetHeaders(c)

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
		true, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"posts":   postsResponse,
		"success": true,
	}

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
		headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}
	response["post"] = fetchPostData

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to edit a Post
func (handlers *FeedHandlers) EditPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")

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
	success := validatePostAttachments(c, editPostRequest.Attachments)
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

		topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, true)
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

	// update post data using helper method
	err = handlers.postHelper.EditPostHelper(postData.ID, editPostRequest.Text, editPostRequest.Heading, editPostRequest.Attachments,
		topicIDs)
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
	fetchPostData, err := fetchPostData(handlers, postId, communityId, commentFilterOptions,
		headers[utils.HeadersMemberId], editPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update post data in elastic search
	err = handlers.esHelper.UpdateDocument(c, ParsePostIndexData(postData), postData.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"post":    fetchPostData,
	})

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

	// remove activity for the post
	deleteActivityFilter := gin.H{
		"entity_type": constants.Post,
		"entity_id":   postData.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

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
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])

	// return final response
	c.JSON(http.StatusOK, parseFetchMultiplePostResponse(handlers.postHelper, createdPostResponse,
		postsCount))
}

func processPostSearchData(handlers *FeedHandlers, data map[string]interface{}, userId string,
	isCm bool, versionCode string, platformCode string) []requests.PostResponse {
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
		handlers.saveHelper, handlers.topicHelper, postList, userId, isCm, versionCode, platformCode)

	return postResponse
}

// Exposed Method to search Posts
func (handlers *FeedHandlers) SearchPost(c *gin.Context) {
	// fetch query params and headers
	headers := utils.GetHeaders(c)
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
		searchPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"posts":   finalResponse,
	})
}

// Exposed Method to search user created Posts
func (handlers *FeedHandlers) SearchUserCreatedPost(c *gin.Context) {
	// fetch query params and headers
	userId := c.Param("user_id")
	headers := utils.GetHeaders(c)
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
		searchPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"posts":   finalResponse,
	})
}
