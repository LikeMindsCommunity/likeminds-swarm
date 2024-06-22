package cache

// Swarm cache keys
const (
	CommunityConfigurationsKey   = "%d_community_configurations"
	UserConnnectionCacheKey      = "connection_list_%s_%s"
	ConnectionFeedBufferCacheKey = "connection_feed_buffer_%s_%s"
	InferdoApiFailsCountKey      = "inferdo_api_fails_count_%d"
	CommunityWebhooksCacheKey    = "%s_webhooks"
	PostTopLikedCommentKey       = "%d_%s_top_liked_comments"
	CommunitySettingsCacheKey    = "%d_community_settings" // communityId
)

// Swarm cache keys of personalised feed
const (
	PostsRececnyMetricsKey  = "%d_recency_metric_score"
	PostsLikesMetricsKey    = "%d_likes_metric_score"
	PostsCommentsMetricsKey = "%d_comments_metric_score"
)

// Cache TTLs
const (
	CommunityConfigurationsCacheTTLInHours = 175 // 7 days
	CommunityWebhooksCacheTTTLInHours      = 175 // 7 days
	PostsRecenctCacheTTLInMins             = 5   // 5 Mins
	CommunitySettingsCacheTTL              = 30 * 24
)

// Kettle Cache Keys
const (
	WidgetMetaCacheKeyKettle = "%d_%s_widget_meta" // community_id widget_id
	TopicMetaCacheKeyKettle  = "%d_%s_topic_meta"  // community_id topic_id
	UserTopicsCacheKeyKettle = "%d_%s_user_topics" // community_id user_id
)
