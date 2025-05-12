package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to fetch comments count of a Post
func fetchPostCommentsCount(helper interfaces.CommentHelper, postId string) int {
	// comment filter data
	commentFilterData := gin.H{
		"post_id":    postId,
		"is_deleted": false,
		"level":      constants.CommentBaseLevel,
	}

	// fetch count using helper method
	commentsCount, err := helper.CountCommentHelper(commentFilterData)
	if err != nil {
		logging.Error("Failed to fetch comments count: ", err)
		return 0
	}

	return int(commentsCount)
}

// Internal Method to fetch comments count of multiple posts
func fetchMultiplePostsCommentsCount(helper interfaces.CommentHelper, postIds []primitive.ObjectID) map[primitive.ObjectID]int {

	count := make(map[primitive.ObjectID]int, len(postIds))

	// fetch comments count for each post
	query := []map[string]interface{}{
		{
			"$match": gin.H{
				"post_id": gin.H{
					"$in": postIds,
				},
				"is_deleted": false,
				"level":      constants.CommentBaseLevel,
			},
		},
		{
			"$group": gin.H{
				"_id": "$post_id",
				"count": gin.H{
					"$sum": 1,
				},
			},
		},
	}

	// fetch comments count using helper method
	results, err := helper.AggregateCommentHelper(query)
	if err != nil {
		logging.Error("Failed to fetch comments count: ", err)
		return count
	}

	// parse comments count
	for _, result := range results.([]gin.H) {
		count[result["_id"].(primitive.ObjectID)] = int(result["count"].(int32))
	}

	return count
}

// Internal Method to fetch replies count of a Comment
func fetchCommentRepliesCount(helper interfaces.CommentHelper, commentId string) (int64, error) {
	commentData, err := fetchCommentByIdInternal(helper, commentId, []string{})
	if err != nil {
		return 0, err
	}

	replyFilterData := gin.H{
		"_id": gin.H{
			"$in": commentData.Replies,
		},
		"is_deleted": false,
	}

	// fetch replies count using helper method
	likesCount, err := helper.CountCommentHelper(replyFilterData)
	if err != nil {
		return 0, err
	}

	return likesCount, nil
}

// Internal Method to fetch parent comment of a Comment
func fetchParentComment(helper interfaces.CommentHelper, commentId primitive.ObjectID,
	postId primitive.ObjectID) (*entities.Comment, error) {
	// comment filter data
	commentFilterData := gin.H{
		"replies":    commentId,
		"is_deleted": false,
		"post_id":    postId,
	}

	// fetch comment using helper method
	commentResults, err := helper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &commentResults[0], nil
}

// Internal Method to fetch a comment using comment_id
func fetchCommentByIdInternal(helper interfaces.CommentHelper, commentId string, excludedUserIds []string,
) (*entities.Comment, error) {

	// comment filter data
	commentFilterData := gin.H{
		"_id":        commentId,
		"is_deleted": false,
	}

	if len(excludedUserIds) > 0 {
		commentFilterData["user_id"] = gin.H{
			"$nin": excludedUserIds,
		}
	}

	// fetch comment using helper method
	commentResults, err := helper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &commentResults[0], nil
}

// fetchCommentByID | get comment by id
func fetchCommentByID(helper interfaces.CommentHelper, commentId string) (*entities.Comment, error) {
	filter := gin.H{
		"_id": commentId,
	}

	commentResults, err := helper.FindCommentHelper(filter, gin.H{})
	if err != nil {
		return nil, err
	}

	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id")
	}

	return &commentResults[0], nil
}

// Internal Method to fetch multiple comments data using comment_ids
func fetchMultipleCommentsData(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, commentIds []string,
) (map[string]responses.CommentWithParentResponse, error) {

	// convert comment_ids to object ids
	commentObjectIds := helpers.ConvertIdsToObjectIds(commentIds)

	// comment filter data
	commentFilterData := gin.H{
		"_id": gin.H{
			"$in": commentObjectIds,
		},
		"community_id": loggedInUser.CommunityId,
	}

	// fetch comments using helper method
	comments, err := handlers.commentHelper.FindCommentHelper(commentFilterData, nil)
	if err != nil {
		return nil, err
	}

	// Make key value pair map for response, comment_id -> comment
	parsedCommentsResponse := map[string]responses.CommentWithParentResponse{}
	for _, comment := range comments {

		// Parse comment for response
		parseCommentsResponse := parseCommentWithParentResponse(handlers, comment, loggedInUser.UserId, loggedInUser.IsCm,
			loggedInUser.VersionCode, loggedInUser.PlatformCode, loggedInUser.ApiRevampCheckV1, loggedInUser.MemberRole)

		// Add to response map
		parsedCommentsResponse[comment.ID.Hex()] = parseCommentsResponse

	}

	return parsedCommentsResponse, nil
}

// Internal Method to fetch a comment using comment_id and post_id
func fetchComment(helper interfaces.CommentHelper, commentId string, postId string, exlcudedUserIds []string) (*entities.Comment, error) {
	// comment filter data
	commentFilterData := gin.H{
		"_id":        commentId,
		"is_deleted": false,
		"post_id":    postId,
		"user_id": gin.H{
			"$nin": exlcudedUserIds,
		},
	}

	// fetch comment using helper method
	commentResults, err := helper.FindCommentHelper(commentFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of comment
	if len(commentResults) == 0 {
		return nil, fmt.Errorf("invalid comment_id sent")
	}

	return &commentResults[0], nil
}

// Internal Method to parse comment for response
func parseCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper, comment entities.Comment,
	userId string, isCm bool, versionCode string, platformCode string, apiRevampV1Check bool, cacheHelper cache.Helper, memberRole string,
) responses.CommentResponse {

	likesCount := fetchEntityLikesCount(likeHelper, comment.ID.Hex(), constants.CommentEntityType)
	var response responses.CommentResponse

	response.ID = comment.ID
	response.TempID = comment.TempID
	response.Text = comment.Text
	response.Level = comment.Level
	response.CommunityId = comment.CommunityId
	response.Attachments = ParseAttachmentsforResponse(comment.Attachments, apiRevampV1Check)
	response.PostId = comment.PostId
	response.UserId = comment.UserId
	response.UUID = comment.UserId
	response.IsLiked = fetchUserLikedStatusByEntity(likeHelper, comment.ID.Hex(), constants.CommentEntityType, userId)
	response.LikesCount = likesCount
	response.IsDeleted = comment.IsDeleted
	response.IsEdited = comment.IsEdited
	response.MenuItems = []responses.MenuResponse{}

	if memberRole != utils.GuestRole {
		loggedInUser := &LoggedInUserParams{
			UserId:           userId,
			CommunityId:      comment.CommunityId,
			IsCm:             isCm,
			VersionCode:      versionCode,
			PlatformCode:     platformCode,
			ApiRevampCheckV1: apiRevampV1Check,
			MemberRole:       memberRole,
		}

		response.MenuItems = getEntityMenuItems(cacheHelper, loggedInUser, constants.CommentEntityType,
			userId == comment.UserId, false, false, comment.UserId)
	}

	if comment.Level == constants.CommentBaseLevel {
		repliesCount, _ := fetchCommentRepliesCount(commentHelper, comment.ID.Hex())
		response.CommentsCount = int(repliesCount)
	}

	if comment.IsDeleted {
		response.DeleteReason = comment.DeleteReason
		response.DeletedBy = comment.DeletedBy
		response.DeletedByUUID = comment.DeletedBy
	}

	response.CreatedAt = int(comment.CreatedAt.UnixMilli())
	response.UpdatedAt = int(comment.UpdatedAt.UnixMilli())

	// ApiRevampV1Check to remove user_id and community_id from comment
	if apiRevampV1Check {
		response.UserId = ""
		response.CommunityId = 0
	}

	return response
}

// Method to fetch comment response using comment id
func FetchSingleCommentWithParentResponse(handlers *FeedHandlers, commentId string) (*responses.CommentWithParentResponse, error) {

	comment, err := fetchCommentByID(handlers.commentHelper, commentId)
	if err != nil {
		return nil, err
	}

	response := parseCommentWithParentResponse(handlers, *comment, comment.UserId, false, "", "", false, utils.DefaultRole)

	return &response, nil
}

// Internal Method to parse multiple comments for response
func parseMultipleCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper, comments []entities.Comment,
	userId string, isCm bool, versionCode string, platformCode string, apiRevampV1Check bool, cacheHelper cache.Helper, memberRole string,
) []responses.CommentResponse {

	response := []responses.CommentResponse{}
	for _, comment := range comments {
		response = append(response, parseCommentResponse(likeHelper, commentHelper, comment, userId, isCm, versionCode, platformCode,
			apiRevampV1Check, cacheHelper, memberRole))
	}

	return response
}

// Internal method to parse comments for FetchCommentsResponse
func parseCommentWithParentResponse(handlers *FeedHandlers, comment entities.Comment, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool, memberRole string) responses.CommentWithParentResponse {

	fetchCommentResponse := responses.CommentWithParentResponse{
		CommentResponse: parseCommentResponse(handlers.likeHelper, handlers.commentHelper, comment, userId, isCm, versionCode, platformCode,
			apiRevampV1Check, handlers.cacheHelper, memberRole),
	}

	// Fetch parent comment if exists
	if fetchCommentResponse.Level > constants.CommentBaseLevel {
		parentComment, err := fetchParentComment(handlers.commentHelper, comment.ID, comment.PostId)
		if err == nil {
			parentCommentResponse := parseCommentResponse(handlers.likeHelper, handlers.commentHelper, *parentComment, userId, isCm, versionCode, platformCode,
				apiRevampV1Check, handlers.cacheHelper, memberRole)
			fetchCommentResponse.ParentComment = &parentCommentResponse
		}
	}

	return fetchCommentResponse

}

// Internal Method to parse comment data for FetchComment API
func parseFetchCommentResponse(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	rawComment *entities.Comment, replies []responses.CommentResponse, userId string, isCm bool,
	versionCode string, platformCode string, apiRevampV1Check bool, cacheHelper cache.Helper,
	memberRole string) responses.FetchCommentResponse {

	var response responses.FetchCommentResponse

	response.CommentResponse = parseCommentResponse(likeHelper, commentHelper, *rawComment, userId, isCm, versionCode, platformCode,
		apiRevampV1Check, cacheHelper, memberRole)

	if len(replies) > 0 {
		response.Replies = replies
	} else {
		response.Replies = []responses.CommentResponse{}
	}

	if response.CommentResponse.Level > constants.CommentBaseLevel {
		commentData, err := fetchParentComment(commentHelper, rawComment.ID, rawComment.PostId)
		if err == nil {
			parentCommentResponse := parseCommentResponse(likeHelper, commentHelper, *commentData, userId, isCm, versionCode, platformCode,
				apiRevampV1Check, cacheHelper, memberRole)
			response.ParentComment = &parentCommentResponse
		}
	}

	return response
}

// Internal Method to fetch comment data with postId
func fetchCommentData(handlers *FeedHandlers, loggedInUser *LoggedInUserParams, commentId string, postId string,
	filterOptions map[string]interface{}, getPostData bool, excludedUserIds []string,
) (responses.FetchCommentResponse, error) {

	var response responses.FetchCommentResponse
	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId, excludedUserIds)
	if err != nil {
		return response, err
	}

	commentFilterData := gin.H{
		"_id": gin.H{
			"$in": commentData.Replies,
		},
		"is_deleted": false,
		"post_id":    postId,
		"user_id": gin.H{
			"$nin": excludedUserIds,
		},
	}

	// fetch comment replies using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, filterOptions)
	if err != nil {
		return response, err
	}

	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults,
		loggedInUser.UserId, loggedInUser.IsCm, loggedInUser.VersionCode, loggedInUser.PlatformCode, loggedInUser.ApiRevampCheckV1,
		handlers.cacheHelper, loggedInUser.MemberRole)
	fetchCommentResponse := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentData, repliesResponse, loggedInUser.UserId, loggedInUser.IsCm, loggedInUser.VersionCode, loggedInUser.PlatformCode,
		loggedInUser.ApiRevampCheckV1, handlers.cacheHelper, loggedInUser.MemberRole)

	// fetch post data if getPostData is true
	if getPostData {
		postData, err := FetchPostData(handlers.postHelper, postId, commentData.CommunityId, true, []string{})
		if err != nil {
			return response, err
		}
		// Parse post response and append to Comment's post_data
		parsePostResponse := parseSinglePostResponse(handlers, postData, loggedInUser)
		fetchCommentResponse.Post = &parsePostResponse
	}

	return fetchCommentResponse, nil
}

// Exposed Method to fetch comment by comment_id
func (handlers *FeedHandlers) FetchCommentById(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	commentId := c.Param("comment_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	userId := headers[utils.HeadersMemberId]

	if paramIsCm == "true" {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
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

	// fetch comment data
	commentData, err := fetchCommentByIdInternal(handlers.commentHelper, commentId, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	commentFilterData := gin.H{
		"_id": gin.H{
			"$in": commentData.Replies,
		},
		"is_deleted": false,
		"post_id":    commentData.PostId.Hex(),
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, commentFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	repliesResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, headers[utils.HeadersMemberId],
		isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, handlers.cacheHelper,
		utils.DefaultRole)
	fetchCommentResponse := parseFetchCommentResponse(handlers.likeHelper, handlers.commentHelper,
		commentData, repliesResponse, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode],
		headers[utils.HeadersPlatformCode], apiRevampV1Check, handlers.cacheHelper, utils.DefaultRole)

	response := gin.H{
		"success": true,
		"comment": fetchCommentResponse,
	}

	// Parse comments to fetch widget_ids``
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, false, headers[utils.HeadersMemberId])

	// return final response
	utils.GenerateSuccessResponse(c, response)
}

// Exposed Method to fetch multiple comments by comment_ids
func (handlers *FeedHandlers) FetchComments(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]

	isCm := utils.IsCMRole(memberRole)
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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

	// Get Query Params
	paramCommentIds := c.Query("comment_ids")
	paramIsCm, _ := strconv.ParseBool(c.Query("user_is_cm"))

	// Check if user is CM or not
	if !paramIsCm {
		utils.GeneralAPIValidationError(c, utils.NotAuthorizedError)
		return
	}

	// Unmarshal comment_ids
	var commentIds []string
	err := json.Unmarshal([]byte(paramCommentIds), &commentIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Fetch comments using comment_ids
	comments, err := fetchMultipleCommentsData(handlers, loggedInUser, commentIds)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// response data
	response := gin.H{
		"success":  true,
		"comments": comments,
	}

	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, false, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed method to fetch comment by comment_id and post_id
func (handlers *FeedHandlers) FetchComment(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	memberRole := headers[utils.HeadersMemberRole]
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]
	userId := headers[utils.HeadersMemberId]

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

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		utils.GeneralAPIValidationError(c, utils.PostIsHiddenError)
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

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, loggedInUser, commentId, postId, commentFilterOptions, false, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response["comment"] = fetchCommentResponse
	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, false, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

func validateCreateCommentRequest(handlers *FeedHandlers, communityId int, createCommentRequest *requests.CreateCommentRequest,
	apiRevampV1Check bool) error {

	// strip text and check if it is empty
	createCommentRequest.Text = strings.Trim(createCommentRequest.Text, " ")

	if createCommentRequest.Text == "" && len(createCommentRequest.Attachments) == 0 {
		return fmt.Errorf("please send text or an attachment in the comment")
	}

	if len(createCommentRequest.Attachments) > 1 {
		return fmt.Errorf("only one attachment is allowed")
	}

	// Validate attachments for comments
	err := ValidateAndUpdateAttachments(handlers, communityId, enums.EntityTypeComment, createCommentRequest.Attachments, apiRevampV1Check, false, false)
	if err != nil {
		return err
	}

	return nil
}

// Exposed Method to comment on a Post
func (handlers *FeedHandlers) CommentPost(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	userId := headers[utils.HeadersMemberId]
	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]
	apiKey := headers[utils.HeadersApiKey]

	isCm := utils.IsCMRole(memberRole)
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
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

	// If users are tagged in post
	if len(createCommentRequest.UUIDs) > 0 {
		similarUUIDs := utils.GetSimilarBetweenArray(createCommentRequest.UUIDs, blockUserValuesList.BlockedUsers)
		if len(similarUUIDs) > 0 {
			utils.GeneralAPIValidationError(c, utils.BlockedUserTagError)
			return
		}

		similarUUIDs = utils.GetSimilarBetweenArray(createCommentRequest.UUIDs, blockUserValuesList.BlockingUsers)
		if len(similarUUIDs) > 0 {
			utils.GeneralAPIValidationError(c, utils.BlockingUserTagError)
			return
		}
	}

	// check if custom creation timestamp is used
	useCustomCreationTimestamp := createCommentRequest.CreatedAt > 0 &&
		float64(createCommentRequest.CreatedAt) <= float64(time.Now().UnixMilli())

	// validate create comment request
	if err := validateCreateCommentRequest(handlers, communityId, &createCommentRequest, apiRevampV1Check); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, excludedUserIds)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		utils.GeneralAPIValidationError(c, utils.PostIsHiddenError)
		return
	}

	// create comment using the helper method
	commentId, err := handlers.commentHelper.CreateCommentHelper(
		createCommentRequest.Text,
		postData.ID,
		communityId,
		constants.CommentBaseLevel,
		userId,
		createCommentRequest.TempID,
		createCommentRequest.CreatedAt,
		createCommentRequest.Attachments,
	)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, enums.WidgetParentEntityTypeComment, createCommentRequest.Attachments,
		commentId.(primitive.ObjectID).Hex(), communityId, userId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// update comment data using helper method
	if err := handlers.commentHelper.EditCommentHelper(commentId.(primitive.ObjectID), createCommentRequest.Text, updatedAttachments, false); err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// tagging and activity
	taggedMembers := createCommentRequest.UUIDs
	isCreatorTagged := slices.Contains(taggedMembers, postData.UserId)

	if !useCustomCreationTimestamp && !postData.IsHidden {
		// cta data for activity
		ctaData := gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  commentId.(primitive.ObjectID).Hex(),
		}

		// Parallelize tagging activities
		// will only help when more than 1 user is tagged.

		for _, member := range taggedMembers {

			err := handlers.taskDistributor.AsyncCreateActivityAndSendNotification(
				func() (interface{}, error) {
					return handlers.CreateActivity(
						postData.CommunityId,
						[]string{userId},
						member,
						constants.CommentEntity,
						commentId.(primitive.ObjectID),
						postData.UserId,
						constants.TaggedInPostComment,
						ctaData,
						false, false, primitive.NilObjectID, "")
				},
				platformCode,
				versionCode,
			)

			if err != nil {
				logging.Error("Tag activity creation and send notification failed:", err)
				return
			}
		}

		if !isCreatorTagged {
			activityID, err := handlers.CreateActivity(
				postData.CommunityId,
				[]string{userId},
				postData.UserId,
				constants.PostEntity,
				postData.ID,
				postData.UserId,
				constants.CommentOnPost,
				ctaData,
				false, false, commentId.(primitive.ObjectID), "")
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			if activityID != nil {
				handlers.CreateAlsoCommentedActivity(activityID, postData, headers, ctaData)
				if err := handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), platformCode, versionCode); err != nil {
					logging.Error("Failed to enqueue send notification : ", err)
				}
			}
		}

		// trigger comment added webhook
		if err := handlers.taskDistributor.TriggerCommentAddedWebhook(commentId.(primitive.ObjectID).Hex(), apiKey); err != nil {
			logging.Error("Error triggering comment added webhook", err)
		}

		// trigger comment tagged webhook
		if len(taggedMembers) > 0 {
			if err := handlers.taskDistributor.TriggerCommentTaggedWebhook(commentId.(primitive.ObjectID).Hex(), taggedMembers, apiKey); err != nil {
				logging.Error("Error triggering comment tagged webhook", err)
			}
		}
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

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(2)

	utils.SafeGo(func() {
		// fetch comment response data
		defer wg.Done()

		fetchCommentResponse, err := fetchCommentData(handlers, loggedInUser, commentId.(primitive.ObjectID).Hex(), postId,
			commentFilterOptions, false, excludedUserIds)
		if err == nil {
			mu.Lock() // we are using mutex because maps are not thread-safe in go.
			response["comment"] = fetchCommentResponse
			mu.Unlock()
		}
	})

	utils.SafeGo(func() {
		defer wg.Done()

		// Parse comments to fetch widget_ids``
		widgets := getWidgetDataFromFeedResponse(handlers, response, communityId, false, userId)

		mu.Lock() // again, we are using mutex because maps are not thread-safe in go.
		response["widgets"] = widgets
		mu.Unlock()
	})

	wg.Wait()

	// Delete top liked comments data in post from cache
	utils.SafeGo(func() {
		handlers.cacheHelper.Del(fmt.Sprintf(cache.PostTopLikedCommentKey, communityId, postId))
	})

	// return final response
	c.JSON(http.StatusOK, response)
}

func validateEditCommentRequest(handlers *FeedHandlers, communityId int, userId string, editCommentRequest *requests.EditCommentRequest,
	postId string, commentId string, apiRevampV1Check bool) (*entities.Comment, error) {

	// strip text and check if it is empty
	editCommentRequest.Text = strings.Trim(editCommentRequest.Text, " ")

	if editCommentRequest.Text == "" && len(editCommentRequest.Attachments) == 0 {
		return nil, fmt.Errorf("please send text or an attachment in the comment")
	}

	if len(editCommentRequest.Attachments) > 1 {
		return nil, fmt.Errorf("only one attachment is allowed")
	}

	// Validate attachments for comments
	err := ValidateAndUpdateAttachments(handlers, communityId, enums.EntityTypeComment, editCommentRequest.Attachments, apiRevampV1Check, true, false)
	if err != nil {
		return nil, err
	}

	// Check if Post_id is valid
	_, err = FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		return nil, err
	}

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId, []string{})
	if err != nil {
		return nil, err
	}

	// If user is not cm and is not the comment creator
	if !editCommentRequest.UserIsCm && commentData.UserId != userId {
		return nil, err
	}

	return commentData, nil
}

// Exposed Method to edit a comment
func (handlers *FeedHandlers) EditComment(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")

	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]
	isCm := utils.IsCMRole(headers[utils.HeadersMemberRole])
	memberRole := headers[utils.HeadersMemberRole]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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

	// validation of request body
	var editCommentRequest requests.EditCommentRequest
	if err := c.ShouldBindJSON(&editCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validate edit comment request
	commentData, err := validateEditCommentRequest(handlers, communityId, headers[utils.HeadersMemberId], &editCommentRequest, postId, commentId, apiRevampV1Check)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, enums.WidgetParentEntityTypeComment, editCommentRequest.Attachments,
		commentData.ID.Hex(), communityId, userId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// update comment data using helper method
	err = handlers.commentHelper.EditCommentHelper(commentData.ID, editCommentRequest.Text, updatedAttachments, true)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Generate page filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, loggedInUser, commentId, postId, commentFilterOptions, false, []string{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
		"comment": fetchCommentResponse,
	}

	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, false, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

func validateCommentReplyRequest(handlers *FeedHandlers, userId string, communityId int, postId string, commentId string,
	createCommentRequest *requests.CreateCommentRequest, apiRevampV1Check bool, isCm bool,
) (*entities.Comment, *entities.Post, error) {

	// strip text and check if it is empty
	createCommentRequest.Text = strings.Trim(createCommentRequest.Text, " ")

	if createCommentRequest.Text == "" && len(createCommentRequest.Attachments) == 0 {
		return nil, nil, fmt.Errorf("please send text or an attachment in the comment")
	}

	if len(createCommentRequest.Attachments) > 1 {
		return nil, nil, fmt.Errorf("only one attachment is allowed")
	}

	// Validate attachments for comments
	err := ValidateAndUpdateAttachments(handlers, communityId, enums.EntityTypeComment, createCommentRequest.Attachments, apiRevampV1Check, false, false)
	if err != nil {
		return nil, nil, err
	}

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		return nil, nil, err
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		return nil, nil, fmt.Errorf(utils.PostIsHiddenError)
	}

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId, []string{})
	if err != nil {
		return nil, nil, err
	}

	// validation of comment level
	if commentData.Level >= constants.CommentAllowedLevel {
		return nil, nil, fmt.Errorf(constants.CommentAllowedErrorMessage)
	}

	return commentData, postData, nil
}

// Exposed Method to Reply on a Comment
func (handlers *FeedHandlers) ReplyComment(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")
	platformCode := headers[utils.HeadersPlatformCode]
	versionCode := headers[utils.HeadersVersionCode]
	memberRole := headers[utils.HeadersMemberRole]

	isCm := utils.IsCMRole(memberRole)
	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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

	// validation of request body
	var createCommentRequest requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&createCommentRequest); err != nil {
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

	// If users are tagged in post
	if len(createCommentRequest.UUIDs) > 0 {
		similarUUIDs := utils.GetSimilarBetweenArray(createCommentRequest.UUIDs, blockUserValuesList.BlockedUsers)
		if len(similarUUIDs) > 0 {
			utils.GeneralAPIValidationError(c, utils.BlockedUserTagError)
			return
		}

		similarUUIDs = utils.GetSimilarBetweenArray(createCommentRequest.UUIDs, blockUserValuesList.BlockingUsers)
		if len(similarUUIDs) > 0 {
			utils.GeneralAPIValidationError(c, utils.BlockingUserTagError)
			return
		}
	}

	// check if custom creation timestamp is used
	var useCustomCreationTimestamp bool = false
	if createCommentRequest.CreatedAt > 0 &&
		float64(createCommentRequest.CreatedAt) <= float64(time.Now().UnixMilli()) {
		useCustomCreationTimestamp = true
	}

	commentData, postData, err := validateCommentReplyRequest(handlers, userId, communityId, postId, commentId, &createCommentRequest, apiRevampV1Check, isCm)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create comment using the helper method
	newCommentId, err := handlers.commentHelper.CreateCommentHelper(createCommentRequest.Text, commentData.PostId, communityId,
		commentData.Level+1, headers[utils.HeadersMemberId], createCommentRequest.TempID, createCommentRequest.CreatedAt,
		createCommentRequest.Attachments)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// process attachments for widgets
	updatedAttachments, err := ProcessAttachmentsForWidgets(handlers, enums.WidgetParentEntityTypeComment, createCommentRequest.Attachments,
		newCommentId.(primitive.ObjectID).Hex(), communityId, userId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	err = handlers.commentHelper.EditCommentHelper(newCommentId.(primitive.ObjectID), createCommentRequest.Text, updatedAttachments, false)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}
	// comment update data
	commentUpdateData := gin.H{
		"$push": gin.H{
			"replies": newCommentId.(primitive.ObjectID),
		},
	}

	// update post using the helper method
	err = handlers.commentHelper.UpdateCommentByIdHelper(commentData.ID, commentUpdateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	taggedMembers, err := getTaggedUsers(createCommentRequest.Text)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	var isCreatorTagged bool = false

	if !useCustomCreationTimestamp && !postData.IsHidden {

		// cta data for activity
		ctaData := gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     postId,
			"comment_id":  newCommentId.(primitive.ObjectID).Hex(),
		}

		for _, member := range taggedMembers {
			if member == commentData.UserId {
				isCreatorTagged = true
			}

			// create tag activity
			activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]}, member,
				constants.CommentEntity, newCommentId.(primitive.ObjectID), postData.UserId, constants.TaggedInPostComment, ctaData,
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

		if !isCreatorTagged {
			// create comment activity
			activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
				commentData.UserId, constants.CommentEntity, commentData.ID, commentData.UserId, constants.CommentOnComment, ctaData,
				false, false, newCommentId.(primitive.ObjectID), "")
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

		// trigger comment added webhook
		err = handlers.taskDistributor.TriggerCommentAddedWebhook(newCommentId.(primitive.ObjectID).Hex(), headers[utils.HeadersApiKey])
		if err != nil {
			logging.Error("Error triggering comment added webhook", err)
		}

		if len(taggedMembers) > 0 {
			// trigger comment tagged webhook
			err = handlers.taskDistributor.TriggerCommentTaggedWebhook(newCommentId.(primitive.ObjectID).Hex(), taggedMembers, headers[utils.HeadersApiKey])
			if err != nil {
				logging.Error("Error triggering comment tagged webhook", err)
			}
		}
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

	// fetch comment response data
	fetchCommentResponse, err := fetchCommentData(handlers, loggedInUser, newCommentId.(primitive.ObjectID).Hex(),
		postId, commentFilterOptions, false, excludedUserIds)
	if err == nil {
		response["comment"] = fetchCommentResponse
	}

	response["widgets"] = getWidgetDataFromFeedResponse(handlers, response, communityId, false, headers[utils.HeadersMemberId])

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Delete a Comment
func (handlers *FeedHandlers) DeleteComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c, handlers.cacheHelper)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deleteCommentRequest requests.DeleteCommentRequest
	if err := c.ShouldBindJSON(&deleteCommentRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch post data
	postData, err := FetchPostData(handlers.postHelper, postId, communityId, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment data
	commentData, err := fetchComment(handlers.commentHelper, commentId, postId, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of user permission
	if !deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != commentData.UserId {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// comment update data
	commentUpdateData := gin.H{
		"$set": gin.H{
			"is_deleted":    true,
			"delete_reason": deleteCommentRequest.DeleteReason,
			"deleted_by":    headers[utils.HeadersMemberId],
		},
	}

	// update post using the helper method
	err = handlers.commentHelper.UpdateCommentByIdHelper(commentData.ID, commentUpdateData)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// remove user post comment activity
	deleteUserPostCommentActivity(handlers, postData, c, headers)

	// remove other activity for the comment
	deleteActivityFilter := gin.H{
		"entity_type": constants.CommentEntity,
		"entity_id":   commentData.ID,
	}
	handlers.activityHelper.DeleteActivityHelper(deleteActivityFilter)

	// create delete activity if deleted by CM
	if deleteCommentRequest.UserIsCm && headers[utils.HeadersMemberId] != commentData.UserId {
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{headers[utils.HeadersMemberId]},
			commentData.UserId, constants.CommentEntity, commentData.ID, commentData.UserId, constants.CMDeletedComment,
			gin.H{}, false, false, primitive.NilObjectID, "")
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

	// Delete top liked comments data in post from cache
	handlers.cacheHelper.Del(fmt.Sprintf(cache.PostTopLikedCommentKey, postData.CommunityId, postData.ID.Hex()))

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// Exposed Method to fetch user created comments
func (handlers *FeedHandlers) FetchUserComments(c *gin.Context) {

	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := c.Param("user_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	versionCode := headers[utils.HeadersVersionCode]
	platformCode := headers[utils.HeadersPlatformCode]
	memberRole := headers[utils.HeadersMemberRole]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

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

	commentFilterData := map[string]any{
		"user_id":      userId,
		"community_id": communityId,
		"is_deleted":   false,
		"level":        0,
	}

	// filter options
	commentFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comments using helper method
	commentResults, err := handlers.commentHelper.FindCommentHelper(commentFilterData, commentFilterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	fetchCommentsResponse := parseMultipleCommentResponse(handlers.likeHelper, handlers.commentHelper, commentResults, headers[utils.HeadersMemberId],
		isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, handlers.cacheHelper, utils.DefaultRole)

	// Fetch post ids
	var postIds []string
	postIdsDataMap := map[string]responses.PostResponse{}

	for _, commentResponse := range fetchCommentsResponse {

		if _, ok := postIdsDataMap[commentResponse.PostId.Hex()]; !ok {
			postIds = append(postIds, commentResponse.PostId.Hex())
			postIdsDataMap[commentResponse.PostId.Hex()] = responses.PostResponse{}
		}
	}

	postIdsDataMap, err = fetchPostResponseMapFromPostIds(handlers, loggedInUser, postIds)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	finalResponse := gin.H{
		"success":  true,
		"comments": fetchCommentsResponse,
		"posts":    postIdsDataMap,
	}

	finalResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalResponse, communityId)
	finalResponse["reposted_posts"] = getOriginalPostForReposts(handlers, loggedInUser, finalResponse)
	finalResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalResponse, communityId, isCm, headers[utils.HeadersMemberId])

	utils.GenerateSuccessResponse(c, finalResponse)
}

func deleteUserPostCommentActivity(handlers *FeedHandlers, postData *entities.Post, c *gin.Context, headers map[string]string) {

	activityFilterData := gin.H{
		"community_id": postData.CommunityId,
		"entity_type":  constants.PostEntity,
		"entity_id":    postData.ID,
		"action":       constants.CommentOnPost,
	}

	activity, err := handlers.activityHelper.FindActivityHelper(activityFilterData, gin.H{})
	if err != nil {
		return
	}

	if activity == nil {
		return
	}

	actionBy := activity[0].ActionBy
	actionByMetadata := activity[0].ActionByMetadata
	if actionByMetadata == nil {
		actionByMetadata = map[string]entities.ActionByMetadata{}
	}

	// fetch all the comments on the post
	commentFilterData := gin.H{
		"user_id":    headers[utils.HeadersMemberId],
		"post_id":    postData.ID,
		"is_deleted": false,
	}

	sortOptions := gin.H{
		"$sort": gin.H{
			"created_at": -1,
		},
	}

	comments, err := handlers.commentHelper.FindCommentHelper(commentFilterData, sortOptions)
	if err != nil {
		return
	}

	// if user still has comments on the post, do not delete activity
	if len(comments) > 0 {

		// update action_by_metadata
		commentData := comments[0]

		actionByMetadata[headers[utils.HeadersMemberId]] = entities.ActionByMetadata{
			EntityId:  commentData.ID,
			CreatedAt: commentData.CreatedAt,
		}

	} else {

		// remove uuid from like action list
		actionBy = utils.RemoveAllOccurenceStringList(activity[0].ActionBy, headers[utils.HeadersMemberId])

		// remove user's data from action_by_metadata
		delete(actionByMetadata, headers[utils.HeadersMemberId])
	}

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
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// if action by is [], no user comments on post, mark activity as deleted
	if len(actionBy) == 0 {
		handlers.activityHelper.DeleteActivityHelper(activityFilterData)
		handlers.activityHelper.WarmupUserActivityFeedCache(activity[0].CommunityID, activity[0].ActionOn)
	}
}

// Internal method to create mongo query to get top comments based on likes
func createTopCommentsBasedOnLikesQuery(postIds []primitive.ObjectID, sortOrder int, commentsCount interface{}, excludedUserIds []string) []map[string]interface{} {
	commentsFilterData := []map[string]interface{}{}

	// Add match logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$match": gin.H{
			"$and": []gin.H{
				{
					"post_id": gin.H{
						"$in": postIds,
					},
				},
				{
					"user_id": gin.H{
						"$nin": excludedUserIds,
					},
				},
				{
					"level": gin.H{
						"$eq": 0,
					},
				},
				{
					"is_deleted": gin.H{
						"$eq": false,
					},
				},
			},
		},
	})

	// Add lookup logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$lookup": gin.H{
			"from": "like",
			"let": gin.H{
				"commentId": "$_id",
			},
			"pipeline": []gin.H{
				{
					"$match": gin.H{
						"$expr": gin.H{
							"$eq": []string{"$entity_id", "$$commentId"},
						},
						"is_deleted": false,
					},
				},
			},
			"as": "likes_data",
		},
	})

	commentsFilterData = append(commentsFilterData, gin.H{
		"$project": gin.H{
			"post_id":    1,
			"comment_id": "$_id",
			"created_at": 1,
			"updated_at": 1,
			"likes_count": gin.H{
				"$size": "$likes_data",
			},
		},
	})

	// Add sort logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$sort": gin.H{
			"likes_count": sortOrder,
			"created_at":  -1,
		},
	})

	// Add group logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$group": gin.H{
			"_id": "$post_id",
			"filtered_comments": gin.H{
				"$push": "$$ROOT",
			},
		},
	})

	// Add project logic
	commentsFilterData = append(commentsFilterData, gin.H{
		"$project": gin.H{
			"_id": 1,
			"top_comments": gin.H{
				"$slice": []interface{}{"$filtered_comments", commentsCount},
			},
		},
	})

	return commentsFilterData
}

// Internal Method to get top n comments against posts based on sorting key, sort order
func getTopCommentsAgainstPostsOnLikes(handlers *FeedHandlers, postIds []primitive.ObjectID, sortOrder int,
	commentsCount interface{}, userId string, communityId int, memberRole string,
) (map[string]interface{}, []string, error) {

	postsTopComments := map[string]interface{}{}
	allCommentsIds := []string{}

	// Get users list who are blocked by userId or blocked the userId
	blockUserValuesList, err := externalHelpers.GetUserBlockList(handlers.cacheHelper, userId, communityId)
	if err != nil {
		return nil, nil, err
	}

	// Combine the above two lists to get excluded user lists
	excludedUserIds := append(blockUserValuesList.BlockedUsers, blockUserValuesList.BlockingUsers...)

	commentsFilterData := createTopCommentsBasedOnLikesQuery(postIds, sortOrder, commentsCount, excludedUserIds)

	// fetch post using helper method
	commentResults, err := handlers.commentHelper.AggregateTopCommentsHelper(commentsFilterData)

	if err != nil {
		return nil, nil, err
	}

	for _, commentResult := range commentResults {
		var topCommentsList []string

		for _, topComment := range commentResult.TopComments {
			topCommentsList = append(topCommentsList, topComment.CommentID.Hex())
			allCommentsIds = append(allCommentsIds, topComment.CommentID.Hex())
		}

		postsTopComments[commentResult.PostID.Hex()] = topCommentsList
	}

	if memberRole != utils.GuestRole {
		// Save data in cache
		for _, postId := range postIds {
			var likedCommentsCacheData []byte
			topCommentsData, ok := postsTopComments[postId.Hex()]

			if ok {
				likedCommentsCacheData, _ = json.Marshal(topCommentsData)
			}

			if likedCommentsCacheData == nil {
				likedCommentsCacheData, _ = json.Marshal([]string{})
			}

			handlers.cacheHelper.Set(fmt.Sprintf(cache.PostTopLikedCommentKey, communityId, postId.Hex()), likedCommentsCacheData, 0)
		}
	}

	return postsTopComments, allCommentsIds, nil
}

// Get top liked comments against posts from cache
func getTopCommentsAgainstPostsOnLikesFromCache(handlers *FeedHandlers, postIds []primitive.ObjectID, communityId int) (map[string]interface{}, []string, bool, error) {
	postsTopComments := map[string]interface{}{}
	allCommentsIds := []string{}

	for _, postId := range postIds {
		var topCommentsList []string

		postTopComments := handlers.cacheHelper.Get(fmt.Sprintf(cache.PostTopLikedCommentKey, communityId, postId.Hex()))

		if postTopComments.Val() == "" {
			return nil, nil, false, nil
		}

		topCommentsBytes, err := postTopComments.Bytes()

		if err != nil {
			return nil, nil, false, err
		}

		json.Unmarshal(topCommentsBytes, &topCommentsList)

		postsTopComments[postId.Hex()] = topCommentsList
		allCommentsIds = append(allCommentsIds, topCommentsList...)
	}

	return postsTopComments, allCommentsIds, true, nil
}

func getTopCommentsAgainstPostsSortOnLikes(handlers *FeedHandlers, loggedInUser *LoggedInUserParams,
	postsResponse []responses.PostResponse, commentSortOrderVal int, commentCount int,
) ([]responses.PostResponse, map[string]responses.CommentWithParentResponse, error) {

	var updatedPostsWithComments []responses.PostResponse
	var postIds []primitive.ObjectID
	var topCommentsAgainstPostsData map[string]interface{}
	var allCommentIds []string
	var allPostsFetched bool = false
	var err error

	for _, postData := range postsResponse {
		postIds = append(postIds, postData.ID)
	}

	if loggedInUser.MemberRole != utils.GuestRole {
		topCommentsAgainstPostsData, allCommentIds, allPostsFetched, err = getTopCommentsAgainstPostsOnLikesFromCache(handlers, postIds, loggedInUser.CommunityId)
	}

	if !allPostsFetched {
		topCommentsAgainstPostsData, allCommentIds, err = getTopCommentsAgainstPostsOnLikes(handlers, postIds, commentSortOrderVal, commentCount, loggedInUser.UserId, loggedInUser.CommunityId, loggedInUser.MemberRole)
	}

	if err != nil {
		return nil, nil, err
	}

	for _, postData := range postsResponse {
		postDataAddress := &postData
		commentsData, _ := topCommentsAgainstPostsData[postData.ID.Hex()].([]string)

		(*postDataAddress).CommentIDs = commentsData
		updatedPostsWithComments = append(updatedPostsWithComments, *postDataAddress)
	}

	filtered_comments, _ := fetchMultipleCommentsData(handlers, loggedInUser, allCommentIds)

	return updatedPostsWithComments, filtered_comments, nil
}

func (handlers *FeedHandlers) CreateAlsoCommentedActivity(activityID interface{}, postData *entities.Post,
	headers map[string]string, ctaData gin.H) {

	postCommentActivity, err := fetchActivity(handlers.activityHelper, activityID.(primitive.ObjectID).Hex())
	if err != nil {
		return
	}

	latestCommentUser := postCommentActivity.ActionBy[len(postCommentActivity.ActionBy)-1]
	previousCommentUsers := utils.RemoveAllOccurenceStringList(postCommentActivity.ActionBy, latestCommentUser)

	// if previousCommentUsers = [], no need to create activity
	if len(previousCommentUsers) == 0 {
		return
	}

	for _, previousCommentUser := range previousCommentUsers {

		// create also commented activity
		activityID, err := handlers.CreateActivity(postData.CommunityId, []string{latestCommentUser}, previousCommentUser,
			constants.PostEntity, postData.ID, postData.UserId, constants.AlsoCommentOnPost, ctaData, false, false,
			primitive.NilObjectID, "")
		if err != nil {
			return
		}

		if activityID != nil {
			err = handlers.taskDistributor.AsyncSendNotification(activityID.(primitive.ObjectID), headers[utils.HeadersPlatformCode], headers[utils.HeadersVersionCode])
			if err != nil {
				logging.Error("Failed to enqueue send notification : ", err)
			}
		}
	}
}
