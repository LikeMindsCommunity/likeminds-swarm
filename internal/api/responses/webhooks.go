package responses

import (
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
)

// WebhookPayload | Webhook payload structure
type WebhookPayload struct {
	Event     string                 `json:"event"`
	CreatedAt int64                  `json:"created_at"`
	ID        string                 `json:"id"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// WebhookPostPayload | Webhook post payload structure
type WebhookPostPayload struct {
	Post          requests.PostResponse              `json:"post"`
	Topics        map[string]requests.TopicResponse  `json:"topics"`
	Widgets       map[string]requests.WidgetResponse `json:"widgets"`
	RepostedPosts map[string]requests.PostResponse   `json:"reposted_posts"`
	PostCreator   externalHelpers.MemberMeta         `json:"post_creator"`
}
