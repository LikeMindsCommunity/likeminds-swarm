package enums

// Webhook types
const (
	PostCreatedWebhookType = "post.created"
	PostLikedWebhookType   = "post.liked"
	PostPinnedWebhookType  = "post.pinned"
	PostTaggedWebhookType  = "post.tagged"
	CommentAddedWebhook    = "comment.added"
	CommentTaggedWebhook   = "comment.tagged"
	CommentReactWebhook    = "comment.react"
)

// Webhook retry limit
const (
	WebhookRetryLimit = 3
)

// Webhook sources
const (
	WebhookSourceLMFeed = "LM_FEED"
)
