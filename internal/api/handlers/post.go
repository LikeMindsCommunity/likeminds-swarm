package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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

// Internal Method to parse post for response
func parsePostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	saveHelper interfaces.SaveHelper, post entities.Post, user_id string, is_cm bool,
	versionCode string, platformCode string) requests.PostResponse {
	likes_count, _ := fetchEntityLikesCount(likeHelper, post.ID.Hex(), constants.PostEntityType)
	replies_count, _ := fetchPostCommentsCount(commentHelper, post.ID.Hex())
	var response requests.PostResponse

	response.ID = post.ID
	response.Text = post.Text
	response.Heading = post.Heading
	response.CommunityId = post.CommunityId
	response.ChatroomId = post.ChatroomId
	response.IsPinned = post.IsPinned
	response.UserId = post.UserId
	response.Attachments = parsePostAttachments(post.Attachments, versionCode, platformCode)
	response.LikesCount = int(likes_count)
	response.CommentsCount = int(replies_count)
	response.IsDeleted = post.IsDeleted
	response.IsEdited = post.IsEdited
	response.IsLiked = fetchUserLikedStatusByEntity(likeHelper, post.ID.Hex(),
		constants.PostEntityType, user_id)
	response.IsSaved = fetchUserSavedStatusByPostId(saveHelper, post.ID.Hex(), user_id)
	response.MenuItems = getEntityMenuItems(constants.PostEntityType, is_cm,
		user_id == post.UserId, post.IsPinned, versionCode, platformCode)

	if post.IsDeleted {
		response.DeleteReason = post.DeleteReason
		response.DeletedBy = post.DeletedBy
	}

	response.CreatedAt = int(post.CreatedAt.UnixMilli())
	response.UpdatedAt = int(post.UpdatedAt.UnixMilli())

	return response
}

// Internal Method to parse multiple post for response
func parseMultiplePostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	saveHelper interfaces.SaveHelper, posts []entities.Post, user_id string, is_cm bool,
	versionCode string, platformCode string) []requests.PostResponse {
	response := []requests.PostResponse{}

	for _, post := range posts {
		response = append(response, parsePostResponse(likeHelper, commentHelper, saveHelper, post,
			user_id, is_cm, versionCode, platformCode))
	}

	return response
}

// Internal Method to parse response for fetch post api
func parseFetchPostResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	parsed_post requests.PostResponse, replies []requests.CommentResponse) requests.FetchPostResponse {
	var response requests.FetchPostResponse

	response.PostResponse = parsed_post

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []requests.CommentResponse{}
	}

	return response
}

// Internal Method to fetch post using post_id and community_id
func fetchPost(helper interfaces.PostHelper, post_id string, community_id int) (*entities.Post, error) {
	// post filter data
	post_filter_data := gin.H{
		"_id":          post_id,
		"is_deleted":   false,
		"community_id": community_id,
	}

	// fetch post using helper method
	post_results, err := helper.FindPostHelper(post_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of post_id
	if len(post_results) == 0 {
		return nil, fmt.Errorf("invalid post_id sent")
	}

	return &post_results[0], nil
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
func fetchPostData(handlers *FeedHandlers, post_id string, community_id int,
	filter_options map[string]interface{}, member_id string, is_cm bool, versionCode string,
	platformCode string) (interface{}, error) {
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		return nil, err
	}

	comment_filter_data := gin.H{
		"level":      constants.CommentBaseLevel,
		"is_deleted": false,
		"post_id":    post_id,
	}

	// fetch comment using helper method
	comment_results, err := handlers.commentHelper.FindCommentHelper(comment_filter_data, filter_options)
	if err != nil {
		return nil, err
	}

	post_response := parsePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, *post_data, member_id, is_cm, versionCode, platformCode)
	replies_response := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper,
		comment_results, member_id, is_cm, versionCode, platformCode)
	fetch_post_response := parseFetchPostResponse(handlers.likeHelper, handlers.commentHelper,
		post_response, replies_response)

	return fetch_post_response, nil
}

// Internal Method to fetch multiple posts data using post_ids
func fetchMultiplePostsData(handlers *FeedHandlers,
	post_ids []string,
	community_id int,
	user_id string,
	is_cm bool,
	versionCode string,
	platformCode string) (map[string]requests.PostResponse, error) {

	// convert post_ids to object ids
	post_object_ids := helpers.ConvertIdsToObjectIds(post_ids)

	// filter options to fetch posts from db
	filter_options := gin.H{
		"_id": gin.H{
			"$in": post_object_ids,
		},
		"community_id": community_id,
	}

	// fetch posts using helper method
	posts_lists, err := handlers.postHelper.FindPostHelper(filter_options, gin.H{})
	if err != nil {
		return nil, err
	}

	// Make key value pair of post_id -> PostResponse
	post_response := map[string]requests.PostResponse{}

	// parse post response data for each post
	for _, post := range posts_lists {
		post_response[post.ID.Hex()] = parsePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper,
			post, user_id, is_cm, versionCode, platformCode)
	}

	return post_response, nil

}

// Exposed Method to create a Post
func (handlers *FeedHandlers) CreatePost(c *gin.Context) {
	// fetch headers
	headers := utils.GetHeaders(c)

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
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

	// create post using the helper method
	post_id, err := handlers.postHelper.CreatePostHelper(createPostRequest.Text,
		createPostRequest.Heading, community_id, headers[utils.HeadersMemberId],
		createPostRequest.Attachments, createPostRequest.ChatroomID)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch post data using new post_id
	post_data, err := fetchPost(handlers.postHelper, post_id.(primitive.ObjectID).Hex(), community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// insert post data in elastic search
	err = handlers.esHelper.InsertDocument(c, ParsePostIndexData(post_data), post_data.ID.Hex(),
		constants.PostIndexName)
	if err != nil {
		log.Print(err.Error())
	}

	// Get tagged members from request
	tagged_members := createPostRequest.UUIDs

	for _, member := range tagged_members {
		// create tag activity
		activityID, err := handlers.CreateActivity(community_id, []string{headers[utils.HeadersMemberId]}, member, constants.Post, post_id.(primitive.ObjectID), headers[utils.HeadersMemberId], constants.TaggedInPost, gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     post_id.(primitive.ObjectID).Hex(),
		}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	// filter options
	filter_options, err := generatePageFilterOptions(c, "")
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch post response data
	fetch_post_data, err := fetchPostData(handlers, post_id.(primitive.ObjectID).Hex(), community_id,
		filter_options, headers[utils.HeadersMemberId], false, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode])
	if err == nil {
		response["post"] = fetch_post_data
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to fetch multiple posts from post_ids
func (handlers *FeedHandlers) FetchPosts(c *gin.Context) {

	// fetch headers
	headers := utils.GetHeaders(c)

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// Get Query Params
	param_post_ids := c.Query("post_ids")
	param_is_cm, _ := strconv.ParseBool(c.Query("user_is_cm"))

	// If user is not cm, return error
	if !param_is_cm {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Unmarshal post_ids
	var post_ids []string
	err := json.Unmarshal([]byte(param_post_ids), &post_ids)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch multiple posts data using internal method
	posts_response, err := fetchMultiplePostsData(handlers, post_ids, community_id, headers[utils.HeadersMemberId],
		true, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"posts":   posts_response,
		"success": true,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to fetch a Post using post_id
func (handlers *FeedHandlers) FetchPost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	param_is_cm := c.Query("user_is_cm")
	is_cm := false

	if param_is_cm == "true" {
		is_cm = true
	}

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// filter options
	comment_filter_options, err := generatePageFilterOptions(c, "")
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// fetch post response data
	fetch_post_data, err := fetchPostData(handlers, post_id, community_id, comment_filter_options,
		headers[utils.HeadersMemberId], is_cm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}
	response["post"] = fetch_post_data

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to edit a Post
func (handlers *FeedHandlers) EditPost(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editPostRequest requests.EditPostRequest
	if err := c.ShouldBindJSON(&editPostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Check if user is cm or post creator
	if !editPostRequest.UserIsCm && post_data.UserId != headers[utils.HeadersMemberId] {
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

	// update post data using helper method
	err = handlers.postHelper.EditPostHelper(post_data.ID, editPostRequest.Text, editPostRequest.Heading, editPostRequest.Attachments)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options
	comment_filter_options, err := generatePageFilterOptions(c, "")
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post response data
	fetch_post_data, err := fetchPostData(handlers, post_id, community_id, comment_filter_options,
		headers[utils.HeadersMemberId], editPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update post data in elastic search
	err = handlers.esHelper.UpdateDocument(c, ParsePostIndexData(post_data), post_data.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"post":    fetch_post_data,
	})

}

// Exposed Method to delete a Post
func (handlers *FeedHandlers) DeletePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deletePostRequest requests.DeletePostRequest
	if err := c.ShouldBindJSON(&deletePostRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deletePostRequest.UserIsCm && headers[utils.HeadersMemberId] != post_data.UserId {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// update data
	update_data := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": deletePostRequest.DeleteReason,
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(post_data.ID, update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// delete post data in elastic search
	err = handlers.esHelper.DeleteDocument(c, post_data.ID.Hex(), constants.PostIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// remove activity for the post
	deleteActivityFilter := gin.H{
		"entity_type": constants.Post,
		"entity_id":   post_data.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// if deleted by CM, create delete activity
	if deletePostRequest.UserIsCm && headers[utils.HeadersMemberId] != post_data.UserId {
		activityID, err := handlers.CreateActivity(post_data.CommunityId, []string{headers[utils.HeadersMemberId]}, post_data.UserId, constants.Post, post_data.ID, post_data.UserId, constants.CMDeletedPost, gin.H{}, false, false)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		SendNotification(activityID.(primitive.ObjectID), *handlers, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// Exposed Method to pin a Post
func (handlers *FeedHandlers) PinPost(c *gin.Context) {
	// fetch url params
	post_id := c.Param("post_id")

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update data
	update_data := gin.H{
		"$set": gin.H{
			"is_pinned": !post_data.IsPinned,
		},
	}

	// update post using the helper method
	err = handlers.postHelper.UpdatePostByIdHelper(post_data.ID, update_data)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch updated post data using post_id
	post_data, err = fetchPost(handlers.postHelper, post_id, community_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// update post data in elastic search
	err = handlers.esHelper.UpdateDocument(c, ParsePostIndexData(post_data), post_data.ID.Hex(),
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
	user_id := c.Param("user_id")
	param_is_cm := c.Query("user_is_cm")
	is_cm := false

	if param_is_cm == "true" {
		is_cm = true
	}

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// post filter data
	post_filter_data := gin.H{
		"user_id":      user_id,
		"is_deleted":   false,
		"community_id": community_id,
	}

	// fetch posts count using helper method
	posts_count, err := handlers.postHelper.CountPostHelper(post_filter_data)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// filter options
	post_filter_options, err := generatePageFilterOptions(c, "")
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post using helper method
	post_results, err := handlers.postHelper.FindPostHelper(post_filter_data, post_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	created_post_response := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, post_results, user_id, is_cm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode])

	// return final response
	c.JSON(http.StatusOK, parseFetchMultiplePostResponse(handlers.postHelper, created_post_response,
		posts_count))
}

func processPostSearchData(handlers *FeedHandlers, data map[string]interface{}, user_id string,
	is_cm bool, versionCode string, platformCode string) []requests.PostResponse {
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

	post_response := parseMultiplePostResponse(handlers.likeHelper, handlers.commentHelper,
		handlers.saveHelper, postList, user_id, is_cm, versionCode, platformCode)

	return post_response
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

	// parsing of chatroom ids
	excluded_chatroom_ids := parseIntArrayParam(searchPostRequest.ExcludedChatroomIDs)
	parsed_excluded_chatroom_ids, _ := json.Marshal(excluded_chatroom_ids)

	// dsl query to search posts
	post_query := GetPostFilterQuery(page, page_size, searchPostRequest.SearchType,
		searchPostRequest.Search, fmt.Sprintf("%v", string(parsed_excluded_chatroom_ids)), community_id)
	response := handlers.esHelper.ExecuteQuery(post_query, constants.PostIndexName)

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
	user_id := c.Param("user_id")
	headers := utils.GetHeaders(c)
	var searchPostRequest requests.SearchPostRequest

	err := c.BindQuery(&searchPostRequest)
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

	if user_id != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// dsl query to search posts
	post_query := GetSelfPostFilterQuery(page, page_size, searchPostRequest.SearchType,
		searchPostRequest.Search, user_id, community_id)
	response := handlers.esHelper.ExecuteQuery(post_query, constants.PostIndexName)

	finalResponse := processPostSearchData(handlers, response, headers[utils.HeadersMemberId],
		searchPostRequest.UserIsCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode])

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"posts":   finalResponse,
	})
}
