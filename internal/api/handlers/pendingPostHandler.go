package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed method to create a pending post for review (similar to Create post method)
func (handlers *FeedHandlers) CreatePendingPostForReview(c *gin.Context) {

	// fetch headers
	headers := utils.GetHeaders(c)

	// Post owner user_id
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
	success := validateAndUpdatePostAttachments(c, handlers, communityId, cppr.Attachments, apiRevampV1Check, false)
	if !success {
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
		postUserId, cppr.Attachments, cppr.ChatroomID, cppr.TempID, topicIDs, "", cppr.Visibility, 0, "under_review")
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// process attachments for widgets
	updatedAttachments, ok := processAttachmentsForWidgets(c, handlers, cppr.Attachments,
		pendingPostId.(primitive.ObjectID).Hex(), communityId, postUserId)
	if !ok {
		return
	}

	// update post data using helper method for widgets
	err = handlers.pendingPostHelper.EditPendingPostHelper(pendingPostId.(primitive.ObjectID), cppr.Text,
		cppr.Heading, updatedAttachments, topicIDs, cppr.Visibility, false, "under_review")
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Call caravan API to create a review report for the pending post

	// Response
	response := gin.H{
		"message": "Pending post created successfully",
	}

	utils.GenereateSuccessResponse(c, response)
}
