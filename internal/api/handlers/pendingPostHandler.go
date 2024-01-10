package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to fetch multiple posts data using post_ids
func fetchMultiplePendingPostsData(handlers *FeedHandlers, pendingPostIds []string, communityId int, userId string,
	isCm bool, versionCode string, platformCode string, apiRevampV1Check bool) (map[string]requests.PostResponse, error) {

	// convert post_ids to object ids
	pendingPostObjectIds := helpers.ConvertIdsToObjectIds(pendingPostIds)

	// filter options to fetch posts from db
	filterOptions := gin.H{
		"_id": gin.H{
			"$in": pendingPostObjectIds,
		},
		"community_id": communityId,
	}

	// fetch posts using helper method
	pendingPostLists, err := handlers.pendingPostHelper.FindPendingPostHelper(filterOptions, gin.H{})
	if err != nil {
		return nil, err
	}

	// Make key value pair of post_id -> PostResponse
	postResponse := map[string]requests.PostResponse{}

	// parse post data for response data for each pending post
	for _, pendingPost := range pendingPostLists {
		postResponse[pendingPost.ID.Hex()] = parsePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper,
			handlers.topicHelper, pendingPost.PostData, userId, isCm, versionCode, platformCode, apiRevampV1Check, handlers.cacheHelper)
	}

	return postResponse, nil

}

// Exposed method to create a pending post for review (similar to Create post method)
func (handlers *FeedHandlers) CreatePendingPostForReview(c *gin.Context) {

	headers := utils.GetHeaders(c)
	postUserId := headers[utils.HeadersMemberId]

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var cppr requests.CreatePostRequest
	if err := c.ShouldBindJSON(&cppr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// strip text to check if it is empty
	cppr.Text = strings.Trim(cppr.Text, " ")

	if cppr.Text == "" && len(cppr.Attachments) == 0 {
		utils.GeneralAPIValidationError(c, "can't create post without content")
		return
	}

	// validation of attachments
	err := validateAndUpdatePostAttachments(handlers, communityId, cppr.Attachments, apiRevampV1Check, false)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If NSFW Filtering is enabled & attachments are present, check for NSFW content and update scores
	if len(cppr.Attachments) > 0 {
		validatePostImagesForNSFWContent(handlers.cacheHelper, postUserId, communityId, &cppr.Attachments, true)
	}

	// convert topic_ids to object ids
	topicIDs := helpers.ConvertIdsToObjectIds(cppr.TopicIds)

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

	// check the visibility of the post
	if cppr.Visibility == "" {
		cppr.Visibility = enums.PublicVisibility
	}

	if cppr.Visibility != enums.PrivateVisibility && cppr.Visibility != enums.PublicVisibility {
		utils.GeneralAPIValidationError(c, "Invalid visibility sent")
		return
	}

	// Create pending post
	pendingPostId, err := handlers.pendingPostHelper.CreatePendingPostHelper(cppr.Text, cppr.Heading, communityId,
		postUserId, cppr.Attachments, cppr.ChatroomID, cppr.TempID, topicIDs, "", cppr.Visibility, cppr.CreatedAt,
		enums.UnderReview)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// process attachments for widgets
	updatedAttachments, err := processAttachmentsForWidgets(handlers, constants.PendingPostEntityType,
		cppr.Attachments, pendingPostId.(primitive.ObjectID).Hex(), communityId, postUserId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// update post data using helper method for widgets
	err = handlers.pendingPostHelper.EditPendingPostHelper(pendingPostId.(primitive.ObjectID), cppr.Text,
		cppr.Heading, updatedAttachments, topicIDs, cppr.Visibility, false, enums.UnderReview)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Call caravan API to create a review report for the pending post
	err = externalHelpers.SendPendingPostForReview(postUserId, communityId, pendingPostId.(primitive.ObjectID).Hex())
	if err != nil {

		// Delete the pending post if there is an error in sending the report
		err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostId.(primitive.ObjectID),
			gin.H{"$set": gin.H{"is_deleted": true}})
		if err != nil {
			// Log the error
			fmt.Println("Error in deleting the pending post: ", pendingPostId.(primitive.ObjectID).Hex(), " after error in sending for review: ", err.Error())
		}

		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Response
	response := gin.H{
		"message": "Pending post created successfully",
	}

	utils.GenereateSuccessResponse(c, response)
}

// Exposed method to approve/reject a pending post under review
func (handlers *FeedHandlers) ApproveOrRejectPendingPost(c *gin.Context) {

	pendingPostId := c.Param("pending_post_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var arpr requests.ApproveRejectPendingPostRequest
	if err := c.ShouldBindJSON(&arpr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of status
	if arpr.Status != enums.Approved && arpr.Status != enums.Rejected {
		utils.GeneralAPIValidationError(c, "Invalid status sent")
		return
	}

	// Get pending post data
	filterData := gin.H{
		"_id":          pendingPostId,
		"community_id": communityId,
		"post_type":    enums.UnderReview,
	}

	pendingPostsData, err := handlers.pendingPostHelper.FindPendingPostHelper(filterData, nil)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(pendingPostsData) == 0 {
		utils.GeneralAPIValidationError(c, "Invalid pending post id sent")
		return
	}

	pendingPostData := pendingPostsData[0]

	updateBody := gin.H{
		"$set": gin.H{
			"status":     arpr.Status,
			"is_deleted": true,
		},
	}

	// If status is approved, call Create post API internally
	if arpr.Status == enums.Approved {

		postData, err := createPostFromPendingPost(handlers, pendingPostData)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		// Send Approval notification
		sendPendingPostApprovalNotification(*handlers, postData.UserId, communityId, postData.ID.Hex())

	} else {

		// Send Rejection notification
		sendPendingPostRejectionNotification(*handlers, pendingPostData.UserId, communityId)

	}

	// Update status of pending post
	err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostData.ID, updateBody)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Generate success response
	utils.GenereateSuccessResponse(c, nil)

}

func createPostFromPendingPost(handlers *FeedHandlers, pendingPostData entities.PendingPost) (*entities.Post, error) {

	// Create attachments
	requestAttachments := []requests.Attachment{}

	// marshal attachments
	bytes, err := json.Marshal(pendingPostData.PostData.Attachments)
	if err != nil {
		return nil, err
	}

	// Unmarshall to Request attachments
	err = json.Unmarshal(bytes, &requestAttachments)
	if err != nil {
		return nil, err
	}

	createPostRequest := requests.CreatePostRequest{
		Text:           pendingPostData.PostData.Text,
		Heading:        pendingPostData.PostData.Heading,
		Attachments:    requestAttachments,
		ChatroomID:     pendingPostData.PostData.ChatroomId,
		ParsedTopicIds: pendingPostData.PostData.TopicIds,
		OriginalAuthor: pendingPostData.PostData.OriginalAuthorUUID,
		Visibility:     pendingPostData.PostData.Visibility,
		TempID:         pendingPostData.PostData.TempId,
	}

	// create post using internal method
	postData, err := createPostAfterValidation(handlers, pendingPostData.UserId, pendingPostData.CommunityID,
		createPostRequest)
	if err != nil {
		return nil, err
	}

	// update all the widgets for the newly created post
	updateWidgetsForNewlyCreatePostFromPendingPost(handlers, postData.ID.Hex(), pendingPostData.CommunityID, pendingPostData.PostData.Attachments)

	return postData, nil
}

// Internal method to update all the widgets for the newly created post
func updateWidgetsForNewlyCreatePostFromPendingPost(handlers *FeedHandlers, postId string, communityId int,
	attachments []entities.Attachment) error {

	for _, attachment := range attachments {

		switch attachment.AttachmentType {
		case enums.CustomWidget, enums.PollWidget, enums.ArticleWidget:

			createdByLm := true

			if attachment.AttachmentType == enums.CustomWidget {
				createdByLm = false
			}

			if attachment.AttachmentMeta.EntityID != primitive.NilObjectID {

				_, err := editWidget(handlers, attachment.AttachmentMeta.EntityID.Hex(), postId, constants.PostEntityType, createdByLm, nil, nil, communityId)
				if err != nil {
					fmt.Println("Error in updating widget for post: ", postId, " and widget: ", attachment.AttachmentMeta.EntityID.Hex())
				}
			}
		}

	}

	return nil
}
