package handlers

import (
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/services/worker/distributor"
)

// Feed Handlers structure for all Helper classes
type FeedHandlers struct {
	likeHelper           interfaces.LikeHelper
	commentHelper        interfaces.CommentHelper
	postHelper           interfaces.PostHelper
	pendingPostHelper    interfaces.PendingPostHelper
	activityHelper       interfaces.ActivityHelper
	saveHelper           interfaces.SaveHelper
	topicHelper          interfaces.TopicHelper
	widgetHelper         interfaces.WidgetHelper
	pollVotesHelper      interfaces.PollVotesHelper
	connectionFeedHelper interfaces.ConnectionFeedHelper
	esHelper             searchElastic.EsHelper
	cacheHelper          cache.Helper
	taskDistributor      distributor.FeedTaskDistributor
	postTopicsHelper     interfaces.PostTopicsHelper
	userTopicsHelper     interfaces.UserTopicsHelper
}

// Exposed Method to get an instance for Feed Handlers
func NewFeedHandlers(
	likeHelper interfaces.LikeHelper,
	commentHelper interfaces.CommentHelper,
	postHelper interfaces.PostHelper,
	pendingPostHelper interfaces.PendingPostHelper,
	saveHelper interfaces.SaveHelper,
	activityHelper interfaces.ActivityHelper,
	topicHelper interfaces.TopicHelper,
	widgetHelper interfaces.WidgetHelper,
	pollVotesHelper interfaces.PollVotesHelper,
	connectionFeedHelper interfaces.ConnectionFeedHelper,
	esHelper searchElastic.EsHelper,
	cacheHelper cache.Helper,
	taskDistributor distributor.FeedTaskDistributor,
	postTopicsHelper interfaces.PostTopicsHelper,
	userTopicsHelper interfaces.UserTopicsHelper) *FeedHandlers {

	return &FeedHandlers{
		likeHelper:           likeHelper,
		commentHelper:        commentHelper,
		postHelper:           postHelper,
		pendingPostHelper:    pendingPostHelper,
		saveHelper:           saveHelper,
		activityHelper:       activityHelper,
		topicHelper:          topicHelper,
		widgetHelper:         widgetHelper,
		pollVotesHelper:      pollVotesHelper,
		connectionFeedHelper: connectionFeedHelper,
		esHelper:             esHelper,
		cacheHelper:          cacheHelper,
		taskDistributor:      taskDistributor,
		postTopicsHelper:     postTopicsHelper,
		userTopicsHelper:     userTopicsHelper,
	}
}
