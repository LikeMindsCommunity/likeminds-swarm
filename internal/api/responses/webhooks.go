package responses

import (
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
)

// Webhook payload structure
type WebhookPayload struct {
	Event     string                 `json:"event"`
	CreatedAt int64                  `json:"created_at"`
	ID        string                 `json:"id"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// Webhook post payload structure
type WebhookPostPayload struct {
	Post        PostResponse                       `json:"post"`
	Topics      map[string]TopicResponse           `json:"topics"`
	Widgets     map[string]requests.WidgetResponse `json:"widgets"`
	PostCreator externalHelpers.MemberMeta         `json:"post_creator"`
}

// Webhook comment payload structure
type WebhookCommentPayload struct {
	Comment        CommentWithParentResponse  `json:"comment"`
	CommentCreator externalHelpers.MemberMeta `json:"comment_creator"`
}
