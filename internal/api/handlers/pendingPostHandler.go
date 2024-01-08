package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	updatedAttachments, err := processAttachmentsForWidgets(handlers, cppr.Attachments, pendingPostId.(primitive.ObjectID).Hex(), communityId, postUserId)
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

	// headers := utils.GetHeaders(c)
	// userID := headers[utils.HeadersMemberId]

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
	if arpr.Status != enums.Accepted && arpr.Status != enums.Rejected {
		utils.GeneralAPIValidationError(c, "Invalid status sent")
		return
	}

	// Fetch pending post documents
	filterData := gin.H{
		"_id": gin.H{
			"$in": helpers.ConvertIdsToObjectIds(arpr.PendingPostIds),
		},
		"community_id": communityId,
	}

	pendingPostsData, err := handlers.pendingPostHelper.FindPendingPostHelper(filterData, nil)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	for _, pendingPostData := range pendingPostsData {

		// If status is approved, call Create post API internally
		if arpr.Status == enums.Accepted {

			err := approvePendingPostAndCallCreatePost(handlers, pendingPostData)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}
		}

		updateBody := gin.H{
			"status": arpr.Status,
		}

		// Update status of pending post
		err = handlers.pendingPostHelper.UpdatePendingPostByIdHelper(pendingPostData.ID, updateBody)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

	}

}

func approvePendingPostAndCallCreatePost(handlers *FeedHandlers, pendingPostData entities.PendingPost) error {

	// Call Create post API internally

	// Delete all the widgets for the pending post

	return nil
}
