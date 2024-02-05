package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

const (
	OrderTypeAscending  int = 1
	OrderTypeDescending int = -1
	OrderTypeDefault    int = 0
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
	taskDistributor      FeedTaskDistributor
}

// Exposed Method to get an instance for Feed Handlers
func NewFeedHandlers(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper, postHelper interfaces.PostHelper,
	pendingPostHelper interfaces.PendingPostHelper, saveHelper interfaces.SaveHelper, activityHelper interfaces.ActivityHelper, topicHelper interfaces.TopicHelper,
	widgetHelper interfaces.WidgetHelper, pollVotesHelper interfaces.PollVotesHelper, connectionFeedHelper interfaces.ConnectionFeedHelper,
	esHelper searchElastic.EsHelper, cacheHelper cache.Helper, taskDistributor FeedTaskDistributor) *FeedHandlers {
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
	}
}

// Internal Method to get pagination params in an API
func fetchPaginationParams(c *gin.Context) (int, int, error) {
	page, err := utils.ParseIntFromQueryParam(c.DefaultQuery("page", "0"), 0)
	if err != nil {
		return 0, 0, err
	}

	page_size, err := utils.ParseIntFromQueryParam(c.DefaultQuery("page_size", "0"), 0)
	if err != nil {
		return 0, 0, err
	}

	if page <= 0 {
		page = 1
	}

	if page_size <= 0 {
		page_size = 10
	}

	if page_size > 100 {
		return page, 0, fmt.Errorf("max page_size limit exceeded")
	}

	return page, page_size, nil
}

// Internal method to add sorting options to a map
func addSortingOptions(options map[string]interface{}, orderBy string, order int) map[string]interface{} {
	if order >= OrderTypeDefault {
		order = OrderTypeAscending
	} else {
		order = OrderTypeDescending
	}

	options["$sort"] = gin.H{
		orderBy: order,
	}

	return options
}

// Internal Method to generate filter from page params from an API
func generatePageFilterOptions(c *gin.Context, sortKeyParam string, sortKeyOrderParam int) (map[string]interface{}, error) {
	// fetch pagination query params
	page, page_size, err := fetchPaginationParams(c)
	if err != nil {
		return nil, err
	}

	// page filter options
	filter_options := gin.H{
		"$skip":  page_size * (page - 1),
		"$limit": page_size,
	}

	sortKey := "created_at"

	if sortKeyParam != "" {
		sortKey = sortKeyParam
	}

	sortKeyOrder := OrderTypeDescending

	if sortKeyOrderParam != OrderTypeDefault {
		sortKeyOrder = sortKeyOrderParam
	}

	filter_options = addSortingOptions(filter_options, sortKey, sortKeyOrder)

	return filter_options, nil
}

// Internal Method to get tagged user Ids from a Text
func getTaggedUsers(text string) ([]string, error) {
	tagged_members := []string{}
	pattern, err := regexp.Compile("route://[member member_profile]+/([a-f0-9]{8}-?[a-f0-9]{4}-?4[a-f0-9]{3}-?[89ab][a-f0-9]{3}-?[a-f0-9]{12})")
	if err != nil {
		return nil, err
	}

	allSubstringMatches := pattern.FindAllStringSubmatch(text, -1)

	for _, occurance := range allSubstringMatches {
		tagged_members = append(tagged_members, occurance[1])
	}

	return tagged_members, nil
}

// Internal Method to fetch menu items for a user on an Entity
func getEntityMenuItems(entity_type string, is_cm bool, is_owner bool, is_pinned bool,
	versionCode string, platformCode string, userId string, communityId int, cacheHelper cache.Helper) []requests.MenuResponse {
	var output_menu_items []requests.MenuResponse
	var externalEntities externalHelpers.ExternalEntities

	isEditEnabled := utils.CheckVersion(utils.EditFeedEntityVersions, versionCode, platformCode)

	switch entity_type {
	case constants.PostEntityType:
		// Get community configurations
		communityConfigurationResponse, _ := externalHelpers.GetCommunityConfigurations(cacheHelper, userId, communityId)

		if communityConfigurationResponse != nil {
			externalEntities = externalHelpers.ExternalEntities{
				communityConfigurationResponse.CommunityConfigurations,
			}
		}

		if is_owner && is_cm {
			output_menu_items = GetIsOwnerIsCmPostMenuItems(is_pinned, isEditEnabled, externalEntities)
		}

		if is_owner && !is_cm {
			output_menu_items = GetIsOwnerNotIsCmPostMenuItems(isEditEnabled, externalEntities)
		}

		if !is_owner && is_cm {
			output_menu_items = GetNotIsOwnerIsCmPostMenuItems(is_pinned, isEditEnabled, externalEntities)
		}

		if !is_owner && !is_cm {
			output_menu_items = GetNotIsOwnerNotIsCmPostMenuItems(isEditEnabled, externalEntities)
		}

	case constants.CommentEntityType:
		if is_owner {
			output_menu_items = GetIsOwnerCommentMenuItems(isEditEnabled, externalEntities)
		}

		if !is_owner && is_cm {
			output_menu_items = GetNotIsOwnerIsCmCommentMenuItems(isEditEnabled, externalEntities)
		}

		if !is_owner && !is_cm {
			output_menu_items = GetNotIsOwnerNotIsCmCommentMenuItems(isEditEnabled, externalEntities)
		}
	}

	return output_menu_items
}

// Internal Method to check if a number lies in Fibonacci Series
func checkIfFibonacciNumber(num int) bool {
	var n3, n1, n2 int = 0, 0, 1

	if num == n1 || num == n2 {
		return true
	}

	n3 = n1 + n2

	for {
		if n3 > num {
			break
		}

		if n3 == num {
			return true
		}
		n1 = n2
		n2 = n3
		n3 = n1 + n2
	}

	return false
}

// Internal Method to parse an Integer Array from query params
func parseIntArrayParam(param string) []int {
	response := []int{}

	intermediate_strings := parseStringArrayParam(param)

	for _, value := range intermediate_strings {
		convertedValue, err := strconv.Atoi(value)
		if err == nil {
			response = append(response, convertedValue)
		}
	}

	return response
}

// Internal Method to parse a String Array from query params
func parseStringArrayParam(param string) []string {
	response := []string{}

	// Removal of square braces from array string
	if len(param) > 0 && param[0] == '[' {
		param = param[1:]
	}

	if len(param) > 0 && param[len(param)-1] == ']' {
		param = param[:len(param)-1]
	}

	// Removal of extra spaces from the array string
	param = strings.TrimSpace(param)

	if len(param) > 0 {
		paramValues := strings.Split(param, ",")

		for _, value := range paramValues {
			value = strings.TrimSpace(value)

			// Removal of quotes from each string from array
			if len(value) > 0 && value[0] == '"' {
				value = value[1:]
			}

			if len(value) > 0 && value[len(value)-1] == '"' {
				value = value[:len(value)-1]
			}

			response = append(response, value)
		}
	}

	return response
}

// creates a task with the provided options and payload and enqueues it
func EnqueueBackgroundTask(client *asynq.Client, taskName string, taskPayload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// adds max retry option
	maxRetry, _ := strconv.Atoi(environment.GoDotEnvVariable("ASYNQ_MAX_RETRY"))
	opts = append(opts, asynq.MaxRetry(maxRetry))

	// creates a new task with the provided options and payload
	task := asynq.NewTask(taskName, taskPayload, opts...)

	// enqueues the task to the queue
	taskInfo, err := client.EnqueueContext(context.Background(), task)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task %w", err)
	}
	return taskInfo, nil
}
