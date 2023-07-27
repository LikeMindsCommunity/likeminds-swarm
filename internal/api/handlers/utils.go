package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Feed Handlers structure for all Helper classes
type FeedHandlers struct {
	likeHelper     interfaces.LikeHelper
	commentHelper  interfaces.CommentHelper
	postHelper     interfaces.PostHelper
	activityHelper interfaces.ActivityHelper
	saveHelper     interfaces.SaveHelper
	topicHelper    interfaces.TopicHelper
	esHelper       searchElastic.EsHelper
}

// Exposed Method to get an instance for Feed Handlers
func NewFeedHandlers(likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper, postHelper interfaces.PostHelper,
	saveHelper interfaces.SaveHelper, activityHelper interfaces.ActivityHelper, topicHelper interfaces.TopicHelper,
	esHelper searchElastic.EsHelper) *FeedHandlers {
	return &FeedHandlers{
		likeHelper:     likeHelper,
		commentHelper:  commentHelper,
		postHelper:     postHelper,
		saveHelper:     saveHelper,
		activityHelper: activityHelper,
		topicHelper:    topicHelper,
		esHelper:       esHelper,
	}
}

// Internal Method to get pagination params in an API
func fetchPaginationParams(c *gin.Context) (int, int, error) {
	// fetch and validate query params
	ParamPage := c.DefaultQuery("page", "0")
	ParamPageSize := c.DefaultQuery("page_size", "0")
	if ParamPage == "" {
		ParamPage = "0"
	}

	if ParamPageSize == "" {
		ParamPageSize = "0"
	}

	page, err := strconv.Atoi(ParamPage)
	if err != nil {
		return 0, 0, err
	}

	page_size, err := strconv.Atoi(ParamPageSize)
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
	if order >= 0 {
		order = 1
	} else {
		order = -1
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

	sortKeyOrder := -1

	if sortKeyOrderParam != 0 {
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
	versionCode string, platformCode string) []requests.MenuResponse {
	var output_menu_items []requests.MenuResponse

	isEditEnabled := utils.CheckVersion(utils.EditFeedEntityVersions, versionCode, platformCode)

	switch entity_type {
	case constants.PostEntityType:
		if is_owner && is_cm {
			output_menu_items = GetIsOwnerIsCmPostMenuItems(is_pinned, isEditEnabled)
		}

		if is_owner && !is_cm {
			output_menu_items = GetIsOwnerNotIsCmPostMenuItems(isEditEnabled)
		}

		if !is_owner && is_cm {
			output_menu_items = GetNotIsOwnerIsCmPostMenuItems(is_pinned, isEditEnabled)
		}

		if !is_owner && !is_cm {
			output_menu_items = GetNotIsOwnerNotIsCmPostMenuItems(isEditEnabled)
		}

	case constants.CommentEntityType:
		if is_owner {
			output_menu_items = GetIsOwnerCommentMenuItems(isEditEnabled)
		}

		if !is_owner && is_cm {
			output_menu_items = GetNotIsOwnerIsCmCommentMenuItems(isEditEnabled)
		}

		if !is_owner && !is_cm {
			output_menu_items = GetNotIsOwnerNotIsCmCommentMenuItems(isEditEnabled)
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
