package handlers

import (
	"fmt"
	"time"

	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
)

// Method to validate attachments for post and comments
func ValidateAndUpdateAttachments(handlers *FeedHandlers, communityId int, entityType string, attachments []requests.Attachment,
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
	case enums.EntityTypePost:
		return validatePostAttachments(communityId, attachments, isEditRequest, isRepost, handlers.widgetHelper)

	case enums.EntityTypeComment:
		return validateCommentAttachments(communityId, attachments, handlers.widgetHelper)

	default:
		return fmt.Errorf("send valid entity_type for attachments")

	}
}

func validateAndUpdateAttachmentsForApiRevamp(attachments []requests.Attachment) error {

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

func validatePostAttachments(communityId int, attachments []requests.Attachment, isEditRequest bool, isRepost bool, widgetHelper interfaces.WidgetHelper,
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

		default:
			return fmt.Errorf("send valid attachment_type in attachment")
		}
	}

	return nil
}

func validateCommentAttachments(communityId int, attachments []requests.Attachment, widgetHelper interfaces.WidgetHelper) error {
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
func validateRepostAttachment(attachment requests.Attachment) (string, bool) {

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
func validatePostAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.EntityID == "" {
		return "send entity_id: <post_id> in attachment_meta", false
	}

	return "", true
}

// Internal Method to validate image attachment
func validateImageAttachment(attachment requests.Attachment) (string, bool) {
	if attachment.AttachmentMeta.Url == "" {
		return "send url in attachment_meta for image", false
	}

	return "", true
}

func validateGIFAttachment(attachment requests.Attachment) error {
	if attachment.AttachmentMeta.Url == "" {
		return fmt.Errorf("send url in attachment_meta for gif")
	}

	return nil
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

// Internal Method to validate custom attachment with context
func validateAndUpdateCustomWidgetAttachment(widgetHelper interfaces.WidgetHelper, attachment requests.Attachment, communityId int,
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
