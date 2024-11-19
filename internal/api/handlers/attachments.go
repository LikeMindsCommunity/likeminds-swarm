package handlers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Method to validate attachments for post and comments
func ValidateAndUpdateAttachments(handlers *FeedHandlers, communityId int, entityType string, attachments []requests.AttachmentRequest,
	apiRevampV1check bool, isEditRequest bool, isRepost bool,
) error {

	// Api revamp check to validate and update attachments
	if apiRevampV1check {
		err := validateAndUpdateAttachmentsForApiRevamp(attachments)
		if err != nil {
			return err
		}
	}

	switch entityType {
	case enums.EntityTypePost, enums.EntityTypePendingPost:
		return validatePostAttachments(communityId, attachments, isEditRequest, isRepost, handlers.widgetHelper)

	case enums.EntityTypeComment:
		return validateCommentAttachments(communityId, attachments, handlers.widgetHelper)

	default:
		return fmt.Errorf("send valid entity_type for attachments")

	}
}

// Internal Method to parse post & comment attachments
func ParseAttachmentsforResponse(attachments []entities.Attachment, apiRevampV1Check bool,
) []responses.AttachmentResponse {

	parsedAttachments := []responses.AttachmentResponse{}

	// Convert attachments to requests.Attachment
	for _, attachment := range attachments {
		attachmentResponse := responses.AttachmentResponse{
			AttachmentType: attachment.AttachmentType,
		}

		if attachment.AttachmentMeta != nil {
			attachmentResponse.AttachmentMeta = &responses.AttachmentMetaResponse{
				Name:                 attachment.AttachmentMeta.Name,
				Url:                  attachment.AttachmentMeta.Url,
				Format:               attachment.AttachmentMeta.Format,
				Size:                 attachment.AttachmentMeta.Size,
				Duration:             attachment.AttachmentMeta.Duration,
				Height:               attachment.AttachmentMeta.Height,
				Width:                attachment.AttachmentMeta.Width,
				PageCount:            attachment.AttachmentMeta.PageCount,
				ThumbnailUrl:         attachment.AttachmentMeta.ThumbnailUrl,
				CoverImageUrl:        attachment.AttachmentMeta.CoverImageUrl,
				Title:                attachment.AttachmentMeta.Title,
				Body:                 attachment.AttachmentMeta.Body,
				ExpiryTime:           attachment.AttachmentMeta.ExpiryTime,
				PollType:             attachment.AttachmentMeta.PollType,
				MultipleSelectState:  attachment.AttachmentMeta.MultipleSelectState,
				MultipleSelectNumber: attachment.AttachmentMeta.MultipleSelectNumber,
				IsAnonymous:          attachment.AttachmentMeta.IsAnonymous,
				AllowAddOption:       attachment.AttachmentMeta.AllowAddOption,
			}

			if attachment.AttachmentMeta.OgTags != nil {
				attachmentResponse.AttachmentMeta.OgTags = &responses.OGTagsResponse{
					Title:       attachment.AttachmentMeta.OgTags.Title,
					Image:       attachment.AttachmentMeta.OgTags.Image,
					Description: attachment.AttachmentMeta.OgTags.Description,
					Url:         attachment.AttachmentMeta.OgTags.Url,
				}
			}

			if attachment.AttachmentMeta.EntityID != primitive.NilObjectID {
				attachmentResponse.AttachmentMeta.EntityID = attachment.AttachmentMeta.EntityID.Hex()
			}
		}

		// Append attachment to attachmentsData
		parsedAttachments = append(parsedAttachments, attachmentResponse)
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

func validateAndUpdateAttachmentsForApiRevamp(attachments []requests.AttachmentRequest) error {

	for i := range attachments {
		// If type in attachments is not empty
		if attachments[i].Type != "" {

			// Check if attachment type is valid
			if !attachments[i].Type.IsValid() {
				return fmt.Errorf("Invalid attachment type: " + attachments[i].Type.ToString())
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
		if err != nil {
			return err
		}
	}

	return nil
}

func validatePostAttachments(communityId int, attachments []requests.AttachmentRequest, isEditRequest bool, isRepost bool, widgetHelper interfaces.WidgetHelper,
) error {

	// validate attachment_meta
	for _, element := range attachments {

		if isRepost {
			errorMessage, ok := validateRepostAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}
			continue
		}

		switch element.AttachmentType {
		case enums.ImageWidget:
			errorMessage, ok := validateImageAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.VideoWidget:
			errorMessage, ok := validateVideoAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.DocumentWidget:
			errorMessage, ok := validateDocumentAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.LinkWidget:
			errorMessage, ok := validateLinkAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.CustomWidget:
			err := validateAndUpdateCustomWidgetAttachment(widgetHelper, element, communityId)
			if err != nil {
				return err
			}

		case enums.PollWidget:
			errorMessage, ok := validatePollAttachment(element, isEditRequest)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.ArticleWidget:
			errorMessage, ok := validateArticleAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.RepostWidget:
			if !isEditRequest {
				return fmt.Errorf("send valid attachment_type in attachment")
			}
		case enums.GIFWidget:
			err := validateGIFAttachment(element)
			if err != nil {
				return err
			}
		case enums.ReelWidget:
			err := validateReelAttachment(element)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("send valid attachment_type in attachment")
		}
	}

	return nil
}

func validateCommentAttachments(communityId int, attachments []requests.AttachmentRequest, widgetHelper interfaces.WidgetHelper) error {
	for _, element := range attachments {
		switch element.AttachmentType {
		case enums.ImageWidget:
			errorMessage, ok := validateImageAttachment(element)
			if !ok {
				return fmt.Errorf(errorMessage)
			}

		case enums.VideoWidget:
			return fmt.Errorf("video is not allowed in comments")

		case enums.DocumentWidget:
			return fmt.Errorf("document is not allowed in comments")

		case enums.LinkWidget:
			return fmt.Errorf("link is not allowed in comments")

		case enums.CustomWidget:
			err := validateAndUpdateCustomWidgetAttachment(widgetHelper, element, communityId)
			if err != nil {
				return err
			}

		case enums.PollWidget:
			return fmt.Errorf("poll is not allowed in comments")

		case enums.ArticleWidget:
			return fmt.Errorf("article is not allowed in comments")

		case enums.RepostWidget:
			return fmt.Errorf("repost is not allowed in comments")

		case enums.GIFWidget:
			err := validateGIFAttachment(element)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("send valid attachment_type in attachment")
		}
	}

	return nil
}

// validateRepostAttachment | validates attachments for a repost
func validateRepostAttachment(attachment requests.AttachmentRequest) (string, bool) {

	switch attachment.AttachmentType {
	case enums.PostWidget:
		errorMessage, ok := validatePostAttachment(attachment)
		if !ok {
			return errorMessage, false
		}
		return "", true
	case enums.ImageWidget:
	case enums.VideoWidget:
	case enums.DocumentWidget:
	case enums.LinkWidget:
	case enums.CustomWidget:
	case enums.PollWidget:
	case enums.ArticleWidget:
	default:
		return "invalid attachment_type in attachment for repost", false
	}

	return "unknown attachment_type in attachment for repost", false
}

// validatePostAttachment | validates post as an attachment for repost
func validatePostAttachment(attachment requests.AttachmentRequest) (string, bool) {
	if attachment.AttachmentMeta.EntityID == "" {
		return "send entity_id: <post_id> in attachment_meta", false
	}

	return "", true
}

// Internal Method to validate image attachment
func validateImageAttachment(attachment requests.AttachmentRequest) (string, bool) {
	if attachment.AttachmentMeta.Url == "" {
		return "send url in attachment_meta for image", false
	}

	return "", true
}

func validateGIFAttachment(attachment requests.AttachmentRequest) error {
	if attachment.AttachmentMeta.Url == "" {
		return fmt.Errorf("send url in attachment_meta for gif")
	}

	return nil
}

// Internal method to validate reel attachment
func validateReelAttachment(attachment requests.AttachmentRequest) error {
	if attachment.AttachmentMeta.Url == "" {
		return fmt.Errorf("send url in attachment_meta for reel")
	}

	return nil
}

// Internal Method to validate video attachment
func validateVideoAttachment(attachment requests.AttachmentRequest) (string, bool) {
	if attachment.AttachmentMeta.Url == "" {
		return "send url in attachment_meta for video", false
	}

	if attachment.AttachmentMeta.Duration == 0 {
		return "send duration in attachment_meta for video", false
	}

	return "", true
}

// Internal Method to validate document attachment
func validateDocumentAttachment(attachment requests.AttachmentRequest) (string, bool) {
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
func validateLinkAttachment(attachment requests.AttachmentRequest) (string, bool) {
	if attachment.AttachmentMeta.OgTags.Url == "" {
		return "send url in og_tags in attachment_meta for link", false
	}

	return "", true
}

// Internal Method to validate custom attachment with context
func validateAndUpdateCustomWidgetAttachment(widgetHelper interfaces.WidgetHelper, attachment requests.AttachmentRequest, communityId int,
) error {

	widgetId := attachment.AttachmentMeta.EntityID
	widgetMeta := attachment.AttachmentMeta.WidgetMeta

	if widgetId == "" && (len(widgetMeta) == 0) {
		return fmt.Errorf("please send entity_id or widget_meta in attachment meta")
	}

	// If widget id is present, validate if widget exists
	if widgetId != "" {
		_, err := fetchWidgetByID(widgetHelper, widgetId, false, communityId)
		if err != nil {
			return err
		}

	}

	return nil
}

// Internal Method to validate poll attachment
func validatePollAttachment(attachment requests.AttachmentRequest, isEditRequest bool) (string, bool) {
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
func validateArticleAttachment(attachment requests.AttachmentRequest) (string, bool) {
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
func processMetaBeforeWidgetCreation(attachment requests.AttachmentRequest, metaData map[string]interface{},
	lmMeta map[string]interface{}, uuid string) (map[string]interface{}, map[string]interface{}, error) {
	switch attachment.AttachmentType {
	case enums.PollWidget:
		// create poll options
		pollOptionObjects, err := createPollOptionObjects(attachment.AttachmentMeta.Options, uuid)
		if err != nil {
			return metaData, lmMeta, err
		}

		lmMeta["options"] = pollOptionObjects
		delete(metaData, "options")

	default:
		if len(lmMeta) == 0 {
			lmMeta = nil
		}
	}

	return metaData, lmMeta, nil
}

// Internal Method to process meta data before widget edition
func processMetaBeforeWidgetEdition(attachment requests.AttachmentRequest, metaData map[string]interface{},
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

// extract repost type attachment from a post
func getRepostWidgetDataFromPost(post *entities.Post) entities.Attachment {
	originalPostAttachments := post.Attachments

	for _, attachment := range originalPostAttachments {
		if attachment.AttachmentType == enums.RepostWidget {
			return attachment
		}
	}
	return entities.Attachment{}
}

// extract post type attachement from a repost
func getPostAttachmentDataFromPost(post entities.Post) entities.Attachment {
	postAttachments := post.Attachments

	for _, attachment := range postAttachments {
		if attachment.AttachmentType == enums.PostWidget {
			return attachment
		}
	}
	return entities.Attachment{}
}

// Internal Method to validate/update post images for NSFW score and return error meta
func validateAndUpdatePostImagesForNSFWContent(cacheHelper cache.Helper, userId string, communityId int,
	attachments *[]requests.AttachmentRequest, existingAttachments *[]entities.Attachment) (gin.H, error) {

	// Check if NSFW Filtering is enabled and API Key is present
	enabled, configuration := externalHelpers.GetNSFWConfigurationsOrDefault(cacheHelper, userId, communityId)

	if enabled && configuration.InferdoApiKey != "" {

		// Get existing image urls from attachments if edit post request
		existingImgUrls := map[string]bool{}
		if existingAttachments != nil && len(*existingAttachments) > 0 {
			for _, attachment := range *existingAttachments {

				if attachment.AttachmentType == enums.ImageWidget && attachment.AttachmentMeta.Url != "" {
					existingImgUrls[attachment.AttachmentMeta.Url] = true
				}
			}
		}

		nsfwImageScores := getNsfwScoresFromImageAttachmentsInParallel(cacheHelper, userId, communityId,
			configuration.InferdoApiKey, *attachments, existingImgUrls)

		nsfwImageIndices := []int{}

		for index, score := range nsfwImageScores {
			if score > configuration.CutoffScore {

				// Append index to nsfwImageIndices
				nsfwImageIndices = append(nsfwImageIndices, index)

				// Update NSFW score in attachment meta
				(*attachments)[index].AttachmentMeta.NsfwScore = score
			}
		}

		// If NSFW images are present, return error message with custom meta
		if len(nsfwImageIndices) > 0 {

			indicesString := ""

			// For all the indices get its ordinal number and append it to the error message
			for i, imageIndex := range nsfwImageIndices {

				indicesString += utils.GetOrdinal(imageIndex + 1)

				if i == len(nsfwImageIndices)-2 {
					indicesString += " and"
				} else if i < len(nsfwImageIndices)-1 {
					indicesString += ","
				}

				indicesString += " "

			}

			errorMessage := fmt.Errorf(fmt.Sprintf(utils.NsfwContentInImageError, indicesString))

			errorMeta := gin.H{
				"title":              "NSFW content detected in images",
				"type":               "nsfw_content_in_image",
				"cta":                "<<route://dialog/nsfw_content>>",
				"nsfw_image_indices": nsfwImageIndices,
			}

			return errorMeta, errorMessage
		}
	}

	return nil, nil
}

// Internal method to fetch NSFW score for images in parallel
func getNsfwScoresFromImageAttachmentsInParallel(cacheHelper cache.Helper, userId string, communityId int,
	inferdoApiKey string, attachments []requests.AttachmentRequest, existingImgUrls map[string]bool) []float64 {

	nsfwScores := make([]float64, len(attachments))

	// Make a channel to receive NSFW scores
	wg := sync.WaitGroup{}

	for index, attachment := range attachments {
		if attachment.AttachmentType == enums.ImageWidget {

			// If image url is already present in existing images, skip it
			if _, exists := existingImgUrls[attachment.AttachmentMeta.Url]; exists {
				continue
			}

			// Increment the WaitGroup counter.
			wg.Add(1)

			// Launch a goroutine with closure to fetch NSFW score for the image and send the index on the channel
			utils.SafeGo(func() {
				func(index int, attachment requests.AttachmentRequest) {

					// Decrement the counter when the goroutine completes.
					defer wg.Done()

					nsfwScore, err := externalHelpers.GetNsfwScoreForImage(cacheHelper, userId, communityId, attachment.AttachmentMeta.Url, inferdoApiKey)

					// if no error and score is greater than 0.0, update the score in the array
					if err == nil && nsfwScore > 0.0 {
						nsfwScores[index] = nsfwScore
					}

				}(index, attachment)
			})
		}
	}

	// wait for all goroutines to complete
	wg.Wait()

	return nsfwScores
}

func validateRepostPostAttachment(postData *entities.Post, editPostRequest requests.EditPostRequest) bool {
	// repost's attached post id should not be updated in edit request
	// repost will have only post type (=8) attachment
	existingOriginalPostID := postData.Attachments[0].AttachmentMeta.EntityID.Hex()

	if len(editPostRequest.Attachments) <= 0 {
		return false
	}
	editRepostRequestPostID := editPostRequest.Attachments[0].AttachmentMeta.EntityID

	return existingOriginalPostID == editRepostRequestPostID
}

// Internal Method to process attachments for widgets
func ProcessAttachmentsForWidgets(handlers *FeedHandlers, parentEntityType string, attachments []requests.AttachmentRequest,
	parentEntityId string, communityId int, userId string) ([]requests.AttachmentRequest, error) {

	updatedAttachments := []requests.AttachmentRequest{}

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
			var entityId string

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
					return nil, err
				}

				// process meta data before widget edition
				updatedMetaData := processMetaBeforeWidgetEdition(attachment, metaData, widgetData.MetaData)

				// update widget from given metadata
				_, err = editWidget(handlers, attachment.AttachmentMeta.EntityID, "", "", true, updatedMetaData, nil, communityId)
				if err != nil {
					return nil, err
				}

				entityId = attachment.AttachmentMeta.EntityID

				// Else create a new LM Created widget
			} else {
				// Generate LM Meta
				lmMeta := map[string]interface{}{}

				// process meta data before widget creation
				metaData, lmMeta, err := processMetaBeforeWidgetCreation(attachment, metaData, lmMeta, userId)
				if err != nil {
					return nil, err
				}

				// create widget from given metadata
				widgetData, err := createWidget(handlers, true, parentEntityId, parentEntityType, metaData, lmMeta, communityId)
				if err != nil {
					return nil, err
				}

				entityId = widgetData.ID.Hex()

			}

			// updated attachment
			updatedAttachment := requests.AttachmentRequest{
				AttachmentType: attachment.AttachmentType,
				AttachmentMeta: requests.AttachmentMetaRequest{
					EntityID: entityId,
				},
			}

			updatedAttachments = append(updatedAttachments, updatedAttachment)

		} else if attachment.AttachmentType == enums.CustomWidget {
			entityId := attachment.AttachmentMeta.EntityID
			widgetMeta := attachment.AttachmentMeta.WidgetMeta

			// If entity id is null and widget meta is present, create a new widget and attach it to post
			if entityId == "" && widgetMeta != nil {

				// create widget from given metadata
				widgetData, err := createWidget(handlers, false, parentEntityId, parentEntityType, widgetMeta, nil, communityId)
				if err != nil {
					return nil, err
				}

				// update attachment with widget id
				attachment = requests.AttachmentRequest{
					AttachmentType: attachment.AttachmentType,
					AttachmentMeta: requests.AttachmentMetaRequest{
						EntityID: widgetData.ID.Hex(),
					},
				}
			}
			updatedAttachments = append(updatedAttachments, attachment)
		} else { // Else do nothing
			updatedAttachments = append(updatedAttachments, attachment)
		}
	}

	return updatedAttachments, nil
}

// Internal Method to parse widget_ids from attachments
func getWidgetIdsFromAttachments(attachments []responses.AttachmentResponse) []primitive.ObjectID {
	widgetIds := map[primitive.ObjectID]bool{}
	finalWidgetIds := []primitive.ObjectID{}

	for _, attachment := range attachments {
		entityId := primitive.NilObjectID
		if attachment.AttachmentMeta != nil {
			entityId, _ = primitive.ObjectIDFromHex(attachment.AttachmentMeta.EntityID)
		} else if attachment.MetaData != nil {
			entityId, _ = primitive.ObjectIDFromHex(attachment.MetaData.EntityID)
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
