package cache

// Swarm cache keys
const (
	CommunityConfigurationsKey    = "%d_community_configurations"   // communityId
	UserConnnectionCacheKey       = "connection_list_%s_%s"         // userId, communityId
	ConnectionFeedBufferCacheKey  = "connection_feed_buffer_%s_%s"  // userId, communityId
	InferdoApiFailsCountKey       = "inferdo_api_fails_count_%d"    // communityId
	CommunityWebhooksCacheKey     = "%s_webhooks"                   // apiKey
	PostTopLikedCommentKey        = "%d_%s_top_liked_comments"      // communityId, postId
	CommunitySettingsCacheKey     = "%d_community_settings"         // communityId
	UserCommunityChannelsCacheKey = "%d_%s_user_community_channels" // communityId, userId
)

// Swarm cache keys of personalised feed
const (
	PostsRecencyMetricsKey   = "%d_recency_metric_score"        // communityId
	PostsLikesMetricsKey     = "%d_likes_metric_score"          // communityId
	PostsCommentsMetricsKey  = "%d_comments_metric_score"       // communityId
	CommunityDefaultFeedKey  = "%d_community_default_feed"      // communityId
	UserPersonalisedFeedKey  = "%d_%s_user_personalised_feed"   // communityId, userId
	UserGroupsMetricsKey     = "%d_%s_user_groups_metric_score" // communityId, userId
	UserTopicsMetricsKey     = "%d_%s_user_topics_metric_score" // communityId, userId
	UserSeenDampenedPostsKey = "%d_%s_seen_dampened_posts"      // communityId, userId
)

// Cache TTLs (in Hours)
const (
	CommunityConfigurationsCacheTTLInHours = 175 // 7 days
	CommunityWebhooksCacheTTTLInHours      = 175 // 7 days
	CommunitySettingsCacheTTL              = 720 // 30 days
	CommunitySettingsCacheTTLInHours       = 720 // 30 days
	UserPersonalisedFeedCacheTTLInHours    = 720 // 30 days
	UserCommunityChannelsCacheTTLInHours   = 720 // 30 days
	UserDampenedPostsCacheTTLInHours       = 720 //30days
)

// Cache TTLs (in Mins)
const (
	PostsRecencyCacheTTLInMins         = 5   // 5 Mins
	DefaultCommunityFeedCacheTTLInMins = 30  // 30 Mins
	UserMetricCacheTTLInHours          = 720 // 30 days
)

// Kettle Cache Keys
const (
	WidgetMetaCacheKeyKettle = "%d_%s_widget_meta" // community_id widget_id
	TopicMetaCacheKeyKettle  = "%d_%s_topic_meta"  // community_id topic_id
	UserTopicsCacheKeyKettle = "%d_%s_user_topics" // community_id user_id
)
